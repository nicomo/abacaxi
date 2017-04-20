package controllers

import (
	"github.com/nicomo/abacaxi/models"
	"github.com/terryh/goisbn"
)

func getIsbnIdentifiers(s string, r *models.Record, idType int) error {
	isbnCleaned, err := goisbn.Cleanup(s)
	if err != nil {
		return err
	}

	// it's a valid isbn
	r.Identifiers = append(r.Identifiers, models.Identifier{Identifier: isbnCleaned, IDType: idType})
	isbnConverted, _ := goisbn.Convert(isbnCleaned)
	if isbnConverted != "" {
		isbnConverted := models.Identifier{Identifier: isbnConverted, IDType: idType}
		r.Identifiers = append(r.Identifiers, isbnConverted)
	}

	return nil
}
