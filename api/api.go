package api

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/bkasin/gogios"
	"github.com/bkasin/gogios/helpers/config"
	"github.com/bkasin/gogios/users"
	"github.com/google/logger"
	"github.com/gorilla/mux"
)

type status struct {
	ID         string
	Title      string
	Status     string
	GoodCount  int
	TotalCount int
}

var primaryDB gogios.Database
var configuration *config.Config
var apiLogger *logger.Logger

// API is the main handler for all API calls
func API(conf *config.Config) {
	// Prepare the web logger
	log, err := os.OpenFile("/var/log/gogios/api.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}
	defer log.Close()

	apiLogger = logger.Init("APILog", conf.Options.Verbose, true, log)
	defer apiLogger.Close()

	primaryDB = conf.Databases[0].Database
	configuration = conf

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/", apiHome)
	secureRouter := router.PathPrefix("/api/auth").Subrouter()
	secureRouter.Use(users.JwtVerify)

	// Check routes
	router.HandleFunc("/api/getAllChecks", getAllChecks)
	router.HandleFunc("/api/getCheck/{check}", getCheckStatus)

	// User routes
	router.HandleFunc("/api/login", apiLogin)
	secureRouter.HandleFunc("/createUser", createNewUser)

	apiLogger.Infoln("Starting API")
	err = http.ListenAndServe(conf.WebOptions.APIIP+":"+strconv.Itoa(conf.WebOptions.APIPort), router)
	if err != nil {
		apiLogger.Fatalln(err.Error())
	}
}

func apiHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "API home page")
	w.WriteHeader(http.StatusTeapot)
}
