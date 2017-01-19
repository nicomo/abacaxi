// Package Sudoc consumes the web services available at http://documentation.abes.fr/sudoc/manuels/administration/aidewebservices/index.html
package sudoc

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/nicomo/EResourcesMetadataHub/logger"
)

type SudocData struct {
	Err  string   `xml:"error"`
	PPNs []string `xml:"query>result>ppn"`
}

func FetchPPN(isbn2ppnURL string) ([]string, error) {

	result := make([]string, 0)

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

	logger.Debug.Printf("%s", b)

	var data SudocData

	if err := xml.Unmarshal(b, &data); err != nil {
		logger.Error.Println(err)
	}

	logger.Debug.Println(data)

	if data.Err != "" {
		dataErr := errors.New(data.Err)
		return result, dataErr
	}

	result = data.PPNs

	return result, nil

}

func SudocGetRecord(recordURL string) (string, error) {
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
