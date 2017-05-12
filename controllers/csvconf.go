package controllers

import (
	"reflect"
	"sort"
	"strconv"

	"github.com/nicomo/abacaxi/models"
)

// csvConf2String returns the csvConf as a string to be displayed in UI
func csvConf2String(c models.TSCSVConf) string {

	var csvConfString string
	sc := csvConfConvert(c)

	// To store the keys in slice in sorted order
	var keys []int
	for k := range sc {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	// construct string with keys in order
	for _, k := range keys {
		if sc[k] != "" {
			csvConfString += sc[k] + "; "
		}
	}

	return csvConfString
}

// csvConfConvert takes a csv configuration, replaces the string keys with int keys using the struct tag (Col1 -> 1)
// Col8 "publishername" becomes 8: "publishername"
func csvConfConvert(c models.TSCSVConf) map[int]string {

	sc := make(map[int]string)
	s := reflect.ValueOf(c)

	for i := 0; i < s.NumField(); i++ {
		vField := s.Field(i)
		tag := s.Type().Field(i).Tag
		if colIndex, err := strconv.Atoi(tag.Get("tag_col")); err == nil {
			sc[colIndex] = vField.Interface().(string)
		}
	}

	return sc
}

// csvConfSwap takes a csv configuration, extracts the values to use as keys, and the types become indexes (Col1 -> 1)
// Col8 "publishername" becomes publishername: 8
func csvConfSwap(c models.TSCSVConf) map[string]int {

	sc := make(map[string]int)
	s := reflect.ValueOf(c)
	for i := 0; i < s.NumField(); i++ {
		vField := s.Field(i)
		tag := s.Type().Field(i).Tag
		if vField.Interface().(string) != "" {
			if colIndex, err := strconv.Atoi(tag.Get("tag_col")); err == nil {
				sc[vField.Interface().(string)] = colIndex
			}
		}
	}
	return sc
}

// csvConfGetNFields returns the number of fields used in a particular TSCVConf struct
func csvConfGetNFields(c models.TSCSVConf) int {
	m := csvConfConvert(c)
	return len(m)
}

// csvConfValidate checks that the required fields are there
func csvConfValidate(c models.TSCSVConf) bool {

	sc := csvConfSwap(c)

	// it's OK to not have any csvConf at all
	if len(sc) == 0 {
		return true
	}

	// if one has a csv Conf, then some fields are required
	_, okTitle := sc["publicationtitle"]
	_, okIDPrint := sc["identifierprint"]
	_, okIDOnline := sc["identifieronline"]
	if (okIDOnline || okIDPrint) && okTitle {
		return true
	}

	return false
}
