package controllers

import (
	"reflect"
	"sort"

	"github.com/nicomo/abacaxi/models"

	"github.com/nicomo/abacaxi/logger"
)

// csvConf2String returns the csvConf as a string to be displayed in UI
func csvConf2String(c map[int]string) string {

	var csvConfString string

	// To store the keys in slice in sorted order
	var keys []int
	for k := range c {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	// To perform the opertion you want
	for _, k := range keys {
		csvConfString += c[k] + "; "
	}

	return csvConfString
}

// csvConfConvert swaps keys and values of the TSCSVConf struct
func csvConfConvert(c models.TSCSVConf) map[int]string {

	sc := make(map[int]string)

	s := reflect.ValueOf(c)
	typeOfc := s.Type()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		typ := f.Kind().String()
		switch typ {
		case "slice": // authors
			for _, v := range f.Interface().([]int) {
				sc[v] = typeOfc.Field(i).Name
			}
		case "int":
			myi := f.Interface().(int)
			if myi == 0 {
				continue
			}
			sc[myi] = typeOfc.Field(i).Name
		}
	}
	return sc
}

func csvConfSwap(c models.TSCSVConf) map[string]int {
	sc := make(map[string]int)

	s := reflect.ValueOf(c).Elem()
	for i := 0; i < s.NumField(); i++ {
		tField := s.Type().Field(i)
		vField := s.Field(i)
		tag := tField.Tag
		logger.Debug.Printf("Field Name: %s,\t Field Value: %v,\t Tag Value: %s\n", tField.Name, vField.Interface(), tag.Get("tag_col"))
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
	/*
		if (c.Isbn == 0 && c.Eisbn == 0) || c.Title == 0 {
			logger.Debug.Println("csvConfValidate false")
			return false
		}
		logger.Debug.Println("csvConfValidate true")*/
	return true
}
