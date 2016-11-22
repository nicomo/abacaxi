// Package models stores the structs for the objects we have
package models

import (
	"time"
)

type Ebook struct {
	DateCreated          time.Time
	DateUpdated          time.Time
	Active               bool
	SfxId                string
	SFXLastHarvest       time.Time
	PublisherLastHarvest time.Time
	SudocLastHarvest     time.Time
	Author               string
	Title                string
	PubYear              string
	Edition              int
	TargetService        string
	OpenURL              string
	Acquired             bool
	Isbns                []Isbn
	Ppns                 []PPN
	MarcRecord           []string
	Deleted              bool
}

type Isbn struct {
	isbn       string
	electronic bool
	primary    bool
}

type PPN struct {
	ppn        string
	electronic bool
	primary    bool
}

//TODO: EbookCreate
func EbookCreate(ebk Ebook) error {
	return nil
}

// TODO: EbookRead
func EbookGetByIsbn(isbn string) (Ebook, error) {
	ebk := new(Ebook{})
	return ebk, nil
}

//TODO: EbookUpdate
func EbookUpdate(ebk Ebook) (Ebook, error) {}

//TODO: EbookSoftDelete
func EbookSoftDelete(ebkId int) error {
	return nil
}

//TODO: EbookDelete
func EbookDelete(ebkId int) error {
	return nil
}
