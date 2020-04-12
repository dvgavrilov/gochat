package route

import (
	"fmt"
	"leto-yanao-1/service/source/config"
	"leto-yanao-1/service/source/logs"
	"leto-yanao-1/service/source/messaging"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
)

var router *chi.Mux

// RegisterRoutes функция
func RegisterRoutes() error {
	router := chi.NewRouter()
	logger := logrus.New()
	logger.Formatter = &logrus.JSONFormatter{
		DisableTimestamp: true,
	}

	router.Use(middleware.Recoverer)
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Use(logs.NewStructuredLogger(logger))
	router.Use(middleware.Logger)

	messaging.Init(router)

	debugStatus, err := strconv.ParseBool(config.MainConfiguration.ApplicationSettings.DebugMode)
	if err != nil {
		log.Println("error with reading config file", err.Error())
	}

	if debugStatus {
		fmt.Println(fmt.Sprintf("start listening http at port %v", config.MainConfiguration.WebSocketSettings.Port))
		http.ListenAndServe(fmt.Sprintf(":%v", config.MainConfiguration.WebSocketSettings.Port), router)
	} else {
		fmt.Println(fmt.Sprintf("start listening https at port %v", config.MainConfiguration.WebSocketSettings.Port))
		// err = http.ListenAndServeTLS(":8090", "cert file", "keyfile")
		// if err != nil {
		// 	return err
		// }
	}

	return nil
}
