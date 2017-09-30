package util

import (
	"log"
	"os"
)

func FindHostname() string {
	name, err := os.Hostname()
	if err != nil {
		log.Fatal("Util:: error while fetching hostname")
	}
	return name
}
