package web

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/bkasin/gogios/helpers"
	"github.com/gorilla/mux"
)

// API is the main handler for all API calls
func API(conf helpers.Config) {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/", apiHome)
	helpers.Log.Fatal(http.ListenAndServe(conf.WebOptions.IP+":"+strconv.Itoa(conf.WebOptions.HTTPPort), router))
}

func apiHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "API home was accessed")
}

func getCheckStatus(w http.ResponseWriter, r *http.Request) {

}
