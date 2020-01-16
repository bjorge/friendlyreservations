package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/beego/goyaml2"
	"github.com/bjorge/friendlyreservations/logger"
)

// Logger is the config logger, set for the implementation
var log = logger.New()

var configMap interface{}

func init() {
	fileName := "config.yaml"

	// when running "go test" get the config from the testdata directory
	if flag.Lookup("test.v") != nil {
		fileName = "testdata/config.yaml"
	}

	// check if the config exists in the local directory
	file, err := os.Open(fileName)
	if err != nil {
		log.LogDebugf("The config file %v does not exist, use environment variables instead", fileName)
		return
	}

	// the config file exists, load it up
	// BUG(bjorge): consider implementing a simple flat yaml reader for the config
	configMap, err = goyaml2.Read(file)
	if err != nil {
		panic(err)
	}

	_, ok := configMap.(map[string]interface{})
	if !ok {
		panic(fmt.Errorf("the config file %v does not contain a map of values", fileName))
	}
	log.LogDebugf("The config file map is: %v", configMap)
}
