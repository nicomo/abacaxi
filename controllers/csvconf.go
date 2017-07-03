package controllers

import (
	"errors"
	"net/http"
	"strconv"
)

func getCSVParams(r *http.Request) (map[string]int, error) {

	csvconf := make(map[string]int)

	for i := 1; i <= 10; i++ {
		v := r.PostFormValue("csvcol" + strconv.Itoa(i))
		if v != "" {
			csvconf[v] = i
		}
	}
	if len(csvconf) == 0 || !csvParamsValidate(csvconf) {
		return nil, errors.New("required fields missing in csv configuration")
	}

	return csvconf, nil
}

// csvParamsValidate checks that the required fields are there
func csvParamsValidate(c map[string]int) bool {
	if (c["identifieronline"] != 0 || c["identifierprint"] != 0) && c["publicationtitle"] != 0 {
		return true
	}
	return false
}
