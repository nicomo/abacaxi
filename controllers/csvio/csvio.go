package csvio

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/nicomo/EResourcesMetadataHub/logger"
)

func CsvProcess() {

	// open csv file
	csvFile, err := os.Open("../../data/cyberlibris_100.csv")
	if err != nil {
		logger.Error.Println("cannot open csv file")
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)

	// we don't know how many fields we have, make it variable
	reader.FieldsPerRecord = -1

	rawCsvData, err := reader.ReadAll()
	if err != nil {
		logger.Error.Println("cannot read content of opened file")
	}

	for _, each := range rawCsvData {
		fmt.Println(each)
	}
}
