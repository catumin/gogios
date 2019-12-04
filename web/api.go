package web

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/bkasin/gogios/helpers"
	"github.com/gorilla/mux"
)

type status struct {
	ID    string `json:"ID"`
	Title string `json:"Title"`
	Good  bool   `json:"good"`
	Asof  string `json:"asof"`
}

var allChecks []status

// API is the main handler for all API calls
func API(conf helpers.Config) {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/", apiHome)
	router.HandleFunc("/api/getAllChecks", getAllChecks)
	router.HandleFunc("/api/getCheck/{check}", getCheckStatus)
	helpers.Log.Fatal(http.ListenAndServe(conf.WebOptions.APIIP+":"+strconv.Itoa(conf.WebOptions.APIPort), router))
}

func getCurrentJSON() []status {
	raw, err := ioutil.ReadFile("/opt/gingertechengine/js/current.json")
	if err != nil {
		helpers.Log.Println("Could not open current.json")
		helpers.Log.Println(err.Error())
	}

	err = json.Unmarshal(raw, &allChecks)

	return allChecks
}

func apiHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "API home page")
	w.WriteHeader(http.StatusTeapot)
}

func getCheckStatus(w http.ResponseWriter, r *http.Request) {
	checkID := mux.Vars(r)["check"]

	allChecks := getCurrentJSON()

	for _, singleCheck := range allChecks {
		if singleCheck.ID == checkID {
			json.NewEncoder(w).Encode(singleCheck)
		}
	}
}

func getAllChecks(w http.ResponseWriter, r *http.Request) {
	allChecks := getCurrentJSON()

	json.NewEncoder(w).Encode(allChecks)
}
