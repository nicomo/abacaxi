// Package sudoc consumes the web services available at http://documentation.abes.fr/sudoc/manuels/administration/aIDewebservices/index.html
package sudoc

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/nicomo/gosudoc"

	"github.com/nicomo/abacaxi/logger"
	"github.com/nicomo/abacaxi/models"
)

// FetchRecord returns a marc record for a given PPN (i.e. sudoc ID for the record)
func FetchRecord(recordURL string) (string, error) {

	var result string

	resp, err := http.Get(recordURL)
	if err != nil {
		logger.Error.Printf("fetch: reading %s %v\n", recordURL, err)
		return result, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		logger.Error.Printf("fetch: reading %s %v\n", recordURL, err)
		return result, err
	}
	result = fmt.Sprintf("%s", b)
	return result, nil
}

// GenChannel creates the initial channel in the Fan out / Fan in process to crawl isbn2PPN web service
func GenChannel(records []models.Record) <-chan models.Record {
	out := make(chan models.Record)
	go func() {
		for _, r := range records {
			out <- r
		}
		close(out)
	}()
	return out
}

// GenI2Input generates the input to be consumed by the sudoc web services
// we use both the ISBNs and ISSNs
func GenI2Input(ri []models.Identifier) []string {
	var s []string
	for _, v := range ri {
		if v.IDType == models.IDTypeOnline || v.IDType == models.IDTypePrint {
			s = append(s, v.Identifier)
			continue
		}
	}
	return s
}

// CrawlPPN takes a channel with a Record, passes it on to gosudoc package, retrieves the result
func CrawlPPN(in <-chan models.Record) <-chan int {
	out := make(chan int)
	go func() {
		for record := range in {
			// generate the url for the web service
			i2input := GenI2Input(record.Identifiers)
			if len(i2input) == 0 {
				logger.Info.Printf("No usable ID in record: %v", record.ID)
				out <- 0
				continue
			}

			// get PPN for input
			var res map[string][]string
			var err error
			if len(i2input[0]) == 8 {
				res, err = gosudoc.Issn2ppn(i2input)
				if err != nil {
					logger.Error.Printf("couldn't get PPN: %v", err)
					out <- 0
					continue
				}
			} else {
				res, err = gosudoc.ID2ppn(i2input, "isbn2ppn")
				if err != nil {
					logger.Error.Printf("couldn't get PPN: %v", err)
					out <- 0
					continue
				}
			}

			// update live record with PPNs
			for _, v := range res {
				for _, ppn := range v {
					var exists bool
					for _, w := range record.Identifiers {
						if ppn == w.Identifier {
							exists = true
							continue
						}
					}
					if !exists {
						newPPN := models.Identifier{Identifier: ppn, IDType: models.IDTypePPN}
						record.Identifiers = append(record.Identifiers, newPPN)
					}
				}
			}

			// update record in DB
			_, err = models.RecordUpdate(record)
			if err != nil {
				logger.Error.Println(err)
				out <- 0
				continue
			}
			// everything OK, notify result channel
			out <- 1

			// as a curtesy to http://www.abes.fr
			time.Sleep(time.Millisecond * 250)
		}
		close(out)
	}()
	return out
}

// CrawlRecords takes a channel with a record, passes it on to FetchRecord, retrieves the result
func CrawlRecords(in <-chan models.Record) <-chan int {

	out := make(chan int)
	go func() {
		for record := range in {
			if err := GetSudocRecord(record); err != nil {
				logger.Error.Printf("failed to get Sudoc Unimarc for record %v: %v", record.ID, err)
				out <- 0
				continue
			}
			// everything OK, notify result channel
			out <- 1
			// as a curtesy to http://www.abes.fr
			time.Sleep(time.Millisecond * 250)
		}
		close(out)
	}()

	return out
}

// MergeResults fans in results from crawlPPN
func MergeResults(cs ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	out := make(chan int)

	// copies values from c to out until c is closed, then calls done
	output := func(c <-chan int) {
		for s := range c {
			out <- s
		}
		wg.Done()
	}

	// number of inbound channels
	wg.Add(len(cs))

	// for each inbound channel, call output
	for _, c := range cs {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// GetSudocRecord tries to retrieve a Unimarc record from Sudoc, in 2 passes :
// get their ID from our IDs
// get the actual record from their ID
func GetSudocRecord(record models.Record) error {

	// Do we already have a Unimarc record ID?
	PPN := record.GetPPN()

	var res map[string][]string // used to store PPN given back by sudoc web service
	var unimarc string          // used to store marc record given back by sudoc web service
	var err error
	if len(PPN) == 0 { // we don't have a PPN yet, let's try to get one

		// take the known identifiers
		input := GenI2Input(record.Identifiers)
		if len(input) == 0 {
			// nothing to work with, can't go any further
			err = errors.New("no usable ID in record")
			return err
		}

		// first ID look like an ISSN, let's try that
		if len(input[0]) == 8 {
			res, err = gosudoc.Issn2ppn(input)
			if err != nil {
				return err
			}
		} else { // we have isbns
			res, err = gosudoc.ID2ppn(input, "isbn2ppn")
			if err != nil {
				return err
			}
		}

		// now we have PPNs, let's insert them into the live record struct
		for _, v := range res {
			if len(v) == 0 { // result empty, abort
				return errors.New("no PPN found")
			}
			for _, value := range v {
				var exists bool
				for _, w := range record.Identifiers {
					if value == w.Identifier {
						exists = true
						continue
					}
				}
				if !exists {
					newPPN := models.Identifier{Identifier: value, IDType: models.IDTypePPN}
					record.Identifiers = append(record.Identifiers, newPPN)
					PPN = append(PPN, newPPN.Identifier)
				}
			}
		}
	}

	// we have a PPN -> now get the unimarc record
	unimarc, err = FetchRecord("http://www.sudoc.fr/" + PPN[0] + ".abes")
	if err != nil {
		return err
	}

	// actually save updated ebook struct to DB
	record.RecordUnimarc = unimarc
	record, err = models.RecordUpdate(record)
	if err != nil {
		return err
	}

	return nil
}

// GetSudocRecords tries to get batches of unimarc record from Sudoc web services
func GetSudocRecords(records []models.Record) {
	// set up the pipeline
	in := GenChannel(records)

	// fan out to 2 workers
	c1 := CrawlRecords(in)
	c2 := CrawlRecords(in)

	// fan in results
	recordsCounter := 0
	for n := range MergeResults(c1, c2) {
		recordsCounter += n
	}

	// let's do a little reporting to the user
	report := models.Report{
		ReportType: models.SudocWs,
	}
	msg := fmt.Sprintf("Number of local records sent : %d - number of unimarc records received  : %d", len(records), recordsCounter)
	report.Text = append(report.Text, msg)
	if recordsCounter == 0 {
		report.Success = false
		report.Text = append(report.Text, "Check the server logs for details.")
		if err := report.ReportCreate(); err != nil {
			logger.Error.Printf("couldn't create report: %v", err)
		}
	}

	report.Success = true
	if err := report.ReportCreate(); err != nil {
		logger.Error.Printf("couldn't create report: %v", err)
	}
}
