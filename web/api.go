package web

import (
	"fmt"
	"net/http"
)

func apiHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "API home was accessed")
}

func getCheckStatus(w http.ResponseWriter, r *http.Request) {
	fmt.Println("")
}
