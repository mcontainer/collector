package util

import (
	"os"
	log "github.com/sirupsen/logrus"
)

const ENDPOINT_KEY = "AGGREGATOR"


func SetAggregatorEndpoint(aggregator *string) {
	endpoint := os.Getenv(ENDPOINT_KEY)
	if endpoint != "" {
		log.WithField("aggregator", endpoint).Info("Find env aggregator variable")
		*aggregator = endpoint
	}
}

func FindHostname() string {
	name, err := os.Hostname()
	if err != nil {
		log.Fatal("Util:: error while fetching hostname")
	}
	return name
}
