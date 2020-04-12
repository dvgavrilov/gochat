package main

import (
	"log"

	"github.com/dvgavrilov/gochat/service/source/config"
	"github.com/dvgavrilov/gochat/service/source/persistence"
	"github.com/dvgavrilov/gochat/service/source/route"
)

func main() {

	err := config.Read()
	if err != nil {
		log.Panic(err.Error())
	}

	err = persistence.Init()
	if err != nil {
		log.Panic(err.Error())
	}
	defer persistence.DataSourceCloser().Close()

	err = route.RegisterRoutes()
	if err != nil {
		log.Panic(err.Error())
	}
}
