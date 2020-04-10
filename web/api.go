package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/bkasin/gogios"
	"github.com/bkasin/gogios/helpers/config"
	"github.com/google/logger"
	"github.com/gorilla/mux"
)

type status struct {
	ID         string
	Title      string
	Status     string
	GoodCount  int
	TotalCount int
	Asof       string
}

var allChecks []status
var primaryDB gogios.Database
var apiLogger *logger.Logger

// API is the main handler for all API calls
func API(conf *config.Config) {
	// Prepare the web logger
	log, err := os.OpenFile("/var/log/gogios/api.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}
	defer log.Close()

	apiLogger := logger.Init("APILog", conf.Options.Verbose, true, log)
	defer apiLogger.Close()

	primaryDB = conf.Databases[0].Database

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/", apiHome)

	router.HandleFunc("/api/getAllChecks", getAllChecks)
	router.HandleFunc("/api/getCheck/{check}", getCheckStatus)

	apiLogger.Infoln("Starting API")
	err = http.ListenAndServe(conf.WebOptions.APIIP+":"+strconv.Itoa(conf.WebOptions.APIPort), router)
	if err != nil {
		apiLogger.Fatalln(err.Error())
	}
}

func getAllStatuses() []status {
	allPrev, err := primaryDB.GetAllCheckRows()
	if err != nil {
		fmt.Printf("Could not read database, error output:\n%s", err.Error())
	}

	var allChecks []status

	for i := 0; i < len(allPrev); i++ {
		if allPrev[i].Model.DeletedAt.String() != "" {
			continue
		}
		allChecks = append(allChecks, status{
			ID:         strconv.FormatUint(uint64(allPrev[i].Model.ID), 10),
			Title:      allPrev[i].Title,
			Status:     allPrev[i].Status,
			GoodCount:  allPrev[i].GoodCount,
			TotalCount: allPrev[i].TotalCount,
			Asof:       allPrev[i].Asof.Format(time.RFC822),
		})
	}

	return allChecks
}

func apiHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "API home page")
	w.WriteHeader(http.StatusTeapot)
}

func getCheckStatus(w http.ResponseWriter, r *http.Request) {
	checkID := mux.Vars(r)["check"]

	allChecks := getAllStatuses()

	for _, singleCheck := range allChecks {
		if singleCheck.ID == checkID {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(singleCheck)
		}
	}
}

func getAllChecks(w http.ResponseWriter, r *http.Request) {
	apiLogger.Infoln("Get All Checks")
	allChecks := getAllStatuses()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allChecks)
}
