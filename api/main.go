package main

import (
	"circulation-supply-api/config"
	"circulation-supply-api/dao"
	"circulation-supply-api/http"
	"flag"
	"log"
)

func init() {
	configFile := flag.String("config", "", "Path to the configuration file")
	flag.Parse()

	var err error
	config.Conf, err = config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	dao.InitDB()
}

func main() {
	http.StartHTTPServer()
}
