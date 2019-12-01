package web

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bkasin/gogios/helpers"
	"github.com/gorilla/mux"
)

type status struct {
	ID    string `json:"ID"`
	Title string `json:"Title"`
	Good  bool   `json:"good"`
}

var allChecks []status

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
	fmt.Fprintf(w, "API home was accessed")
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
