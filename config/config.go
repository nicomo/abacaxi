package config

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/nicomo/abacaxi/logger"
)

// Conf : base configuration information pulled on init from a json file
type Conf struct {
	Hostname         string `json:"hostname"`
	MongoDBHost      string `json:"mongodbhosts"`
	AuthDatabase     string `json:"authdatabase"`
	SessionSecretKey string `json:"sesssecretkey"`
	SessionName      string `json:"sessname"`
	SessionPath      string `json:"sesspath"`
	SessionDomain    string `json:"sessdomain"`
	SessionMaxAge    int    `json:"sessmaxage"`
	SessionSecure    bool   `json:"sesssecure"`
	SessionHttpOnly  bool   `json:"sesshttponly"`
}

// GetConfig generates a Conf object from a json file
func GetConfig() Conf {

	// open & read the json conf file
	file, err := ioutil.ReadFile("config/config.json")
	if err != nil {
		logger.Error.Println(err)
		os.Exit(1)
	}

	// unmarshal json into a config struct
	config := Conf{}
	ErrJSONUnmarshal := json.Unmarshal(file, &config)
	if ErrJSONUnmarshal != nil {
		logger.Error.Println(ErrJSONUnmarshal)
		os.Exit(1)
	}

	return config
}
