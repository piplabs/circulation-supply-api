package main

import (
	"flag"
	"fmt"
	"log"

	"circulation-supply-api/config"
	"circulation-supply-api/dao"
	"circulation-supply-api/metrics"
	"circulation-supply-api/service"

	_ "go.uber.org/automaxprocs"
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

	go metrics.StartMetricsServer(fmt.Sprintf("%s:%d", config.Conf.Metric.Listen, config.Conf.Metric.Port))

	service.Start()
}
