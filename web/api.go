package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bkasin/gogios"
	"github.com/bkasin/gogios/helpers"
	"github.com/bkasin/gogios/helpers/config"
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

// API is the main handler for all API calls
func API(conf *config.Config) {
	primaryDB = conf.Databases[0].Database

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/", apiHome)
	router.HandleFunc("/api/getAllChecks", getAllChecks)
	router.HandleFunc("/api/getCheck/{check}", getCheckStatus)
	helpers.Log.Fatal(http.ListenAndServe(conf.WebOptions.APIIP+":"+strconv.Itoa(conf.WebOptions.APIPort), router))
}

func getAllStatuses() []status {
	allPrev, err := primaryDB.GetAllRows()
	if err != nil {
		helpers.Log.Println("Could not read database")
		helpers.Log.Println(err.Error())
	}

	var allChecks []status

	for i := 0; i < len(allPrev); i++ {
		allChecks = append(allChecks, status{
			ID:         strconv.FormatUint(uint64(allPrev[i].ID), 10),
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
	allChecks := getAllStatuses()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allChecks)
}
