package main

import (
	"leto-yanao-1/service/source/config"
	"leto-yanao-1/service/source/persistence"
	"leto-yanao-1/service/source/route"
	"log"
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
