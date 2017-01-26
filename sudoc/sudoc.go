// Package sudoc consumes the web services available at http://documentation.abes.fr/sudoc/manuels/administration/aIDewebservices/index.html
package sudoc

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/nicomo/abacaxi/logger"
	"github.com/nicomo/abacaxi/models"
)

// PPNData is used to parse xml response
type PPNData struct {
	Err  string   `xml:"error"`
	PPNs []string `xml:"query>result>ppn"`
}

// PPNDataResult is the type returned
type PPNDataResult struct {
	Err  error
	PPNs []string
}

// FetchPPN retrieves ebook ppns from the sudoc web service
func FetchPPN(isbn2ppnURL string) PPNDataResult {

	resp, err := http.Get(isbn2ppnURL)
	if err != nil {
		logger.Error.Printf("fetch: reading %s %v\n", isbn2ppnURL, err)
		panic(err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		logger.Error.Printf("fetch: reading %s: %v\n", isbn2ppnURL, err)
		panic(err)
	}

	var data PPNData
	var result PPNDataResult

	if err := xml.Unmarshal(b, &data); err != nil {
		logger.Error.Println(err)
	}

	if data.Err != "" {
		result.Err = errors.New(data.Err)
		return result
	}

	result.PPNs = data.PPNs

	return result
}

// GetRecord returns a marc record for a given PPN (i.e. sudoc ID for the record)
func GetRecord(recordURL string) (string, error) {

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
func GenChannel(ebks []models.Ebook) <-chan models.Ebook {
	out := make(chan models.Ebook)
	go func() {
		for _, ebk := range ebks {
			out <- ebk
		}
		close(out)
	}()
	return out
}

// GenI2PURL generates an url to be consumed by the isbn2ppn web service
// we place the "electronic" isbns first, then the other isbns
//FIXME: add error management
func GenI2PURL(ebk models.Ebook) string {

	i2purl := "http://www.sudoc.fr/services/isbn2ppn/"
	m := make(map[int]string)
	var se, s []string

	// 2 slices : on for electronic isbns, one for others
	for _, v := range ebk.Isbns {
		if v.Electronic {
			se = append(se, v.Isbn)
			continue
		}
		s = append(s, v.Isbn)
	}

	// we put the electronic isbns in the map first
	for i := 0; i < len(se); i++ {
		m[i] = se[i]
	}

	// then the others
	for i := 0; i < len(s); i++ {
		m[len(se)+i] = s[i]
	}

	// we generate a single url string from the map
	// with e-isbns first
	for i := 0; i < len(m); i++ {
		if i == len(m)-1 {
			i2purl = i2purl + m[i]
			continue
		}
		i2purl = i2purl + m[i] + ","
	}

	return i2purl
}

// CrawlPPN takes a channel with an Ebook, passes it on to FetchPPN, retrieves the result
func CrawlPPN(in <-chan models.Ebook) <-chan int {
	out := make(chan int)
	go func() {
		for ebk := range in {
			// generate the url for the web service
			i2purl := GenI2PURL(ebk)

			// get PPN for i2purl
			result := FetchPPN(i2purl)
			if result.Err != nil {
				logger.Error.Println(result.Err)
				out <- 0
				continue
			}

			// add ppn result in ebk struct
			ebk.Ppns = result.PPNs

			// update record in DB
			// NOTE: would be better to get back to controller and controller calls models.EbookUpdate
			_, err := models.EbookUpdate(ebk)
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

// CrawlRecords takes a channel with an ebook, passes it on to FetchRecord, retrieves the result
func CrawlRecords(in <-chan models.Ebook) <-chan int {
	out := make(chan int)
	go func() {
		for ebk := range in {
			for i := 0; i < len(ebk.Ppns); i++ {
				// generate the URL for the web service
				crurl := "http://www.sudoc.fr/" + ebk.Ppns[i] + ".abes"

				// get record for this PPN
				result, err := GetRecord(crurl)
				if err != nil {
					logger.Error.Println(err)
					continue
				}

				// add record to ebook struct
				ebk.RecordUnimarc = result
				break
			}

			// update record in DB
			_, err := models.EbookUpdate(ebk)
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
