package main

import (
	"circulation-supply-api/api/http"
	"circulation-supply-api/config"
	"circulation-supply-api/dao"
	"flag"
	"log"
)

func main() {
	configFile := flag.String("config", "", "Path to the configuration file")
	flag.Parse()

	var err error
	config.Conf, err = config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	dao.InitDB()

	http.StartHTTPServer()
}
