// Package models stores the structs for the objects we have
package models

import (
	"time"

	"github.com/nicomo/minymapp/logger"
)

type Ebook struct {
	DateCreated          time.Time
	DateUpdated          time.Time
	Active               bool
	SfxId                string
	SFXLastHarvest       time.Time
	PublisherLastHarvest time.Time
	SudocLastHarvest     time.Time
	Authors              []string
	Title                string
	Publisher            string
	Pubdate              string
	Edition              int
	Lang                 string
	TargetService        string
	OpenURL              string
	PackageURL           string
	Acquired             bool
	Isbns                []Isbn
	Ppns                 []PPN
	MarcRecord           []string
	Deleted              bool
}

type Isbn struct {
	Isbn       string
	Electronic bool
	Primary    bool
}

type PPN struct {
	Ppn        string
	Electronic bool
	Primary    bool
}

//TODO: EbookCreate
func EbookCreate(ebk Ebook) error {
	return nil
}

// TODO: EbookRead
func EbookGetByIsbn(isbn string) (Ebook, error) {
	ebk := Ebook{}

	// TODO: 2 types of errors:
	// -- EbookNotFoundErr
	// -- err
	//EbookNotFoundErr := errors.New("we don't have an ebook with this isbn in our records")

	return ebk, nil
}

//TODO: EbooksCreateOrUpdate
func EbooksCreateOrUpdate(records []Ebook) error {
	for _, record := range records { // for each record
		for _, isbn := range record.Isbns { // for each isbn
			workingRecord, err := EbookGetByIsbn(isbn.Isbn) // test if we already know this ebook
			if err != nil {
				/*
					if err == EbookNotFoundErr {
						logger.Error.Println(EbookNotFoundErr) // we don't know the book, let's go create one

					}*/
				logger.Error.Println(err)
				return err
			}
			// we know this isbn already, let's update the ebook
			EbookUpdate(workingRecord)
		}

	}
	return nil
}

//TODO: EbookUpdate
func EbookUpdate(ebk Ebook) (Ebook, error) {
	return ebk, nil
}

//TODO: EbookSoftDelete
func EbookSoftDelete(ebkId int) error {
	return nil
}

//TODO: EbookDelete
func EbookDelete(ebkId int) error {
	return nil
}
