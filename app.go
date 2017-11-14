package main

import (
	"docker-visualizer/collector/cmd"
	log "github.com/sirupsen/logrus"
	"os"
)

var (
	VERSION string
	COMMIT  string
	BRANCH  string
)

func main() {

	if e := cmd.CreateRootCmd("collector", VERSION, COMMIT, BRANCH).Execute(); e != nil {
		log.Error(e)
		os.Exit(1)
	}

}
