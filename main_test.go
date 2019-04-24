package main

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

func init() {
	log.SetOutput(ioutil.Discard)
}
