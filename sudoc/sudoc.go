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

	"github.com/nicomo/EResourcesMetadataHub/logger"
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

// GenChannel creates the initial channel in the Fan in/out process to crawl isbn2PPN web service
func GenChannel(urls []string) <-chan string {
	out := make(chan string)
	go func() {
		for _, s := range urls {
			out <- s
		}
		close(out)
	}()
	return out
}

// CrawlPPN takes a channel with a url as string, pass it on to FetchPPN, retrieves the result
func CrawlPPN(in <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		for s := range in {
			result := FetchPPN(s)
			if result.Err != nil {
				out <- "error for " + s + "\n"
				continue
			}
			out <- "ok for: " + s + "\n"
			time.Sleep(time.Millisecond * 250)
		}
		close(out)
	}()
	return out
}

// MergePPN fans in results from crawlPPN
func MergePPN(cs ...<-chan string) <-chan string {
	var wg sync.WaitGroup
	out := make(chan string)

	// copies values from c to out until c is closed, then calls done
	output := func(c <-chan string) {
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
