package util

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strconv"
)

const (
	ENDPOINT_KEY = "AGGREGATOR"
)

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

func IsRoot() bool {
	cmd := exec.Command("id", "-u")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	i, err := strconv.Atoi(string(output[:len(output)-1]))
	if err != nil {
		log.Fatal(err)
	}
	return i == 0
}
