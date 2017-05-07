package controllers

import (
	"io"
	"net/http"
	"os"

	"github.com/nicomo/abacaxi/config"
	"github.com/nicomo/abacaxi/logger"
)

func exportFile(w http.ResponseWriter, r *http.Request, filename string, filesize int64) error {

	// create a client with hostname from configuration
	conf := config.GetConfig()
	client := &http.Client{}

	resp, err := client.Get(conf.Hostname + "/static/downloads/" + filename)
	if err != nil {
		logger.Error.Println(err)
		return err
	}
	defer resp.Body.Close()

	//stream the body, which is a Reader, to the client without fully loading it into memory
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	written, err := io.Copy(w, resp.Body)
	if err != nil {
		logger.Error.Println(err)
		return err
	}

	// make sure download went OK, then delete file on server
	if filesize == written {
		ErrFDelete := os.Remove("./static/downloads/" + filename)
		if ErrFDelete != nil {
			logger.Error.Println(ErrFDelete)
		}
	}
	return nil
}
