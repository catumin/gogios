package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

var allChecks []status

func getAllStatuses() []status {
	allPrev, err := primaryDB.GetAllChecks()
	if err != nil {
		apiLogger.Errorf("Could not read database, error output:\n%s", err.Error())
	}

	var allChecks []status

	for i := 0; i < len(allPrev); i++ {
		allChecks = append(allChecks, status{
			ID:         strconv.FormatUint(uint64(allPrev[i].Model.ID), 10),
			Title:      allPrev[i].Title,
			Status:     allPrev[i].Status,
			GoodCount:  allPrev[i].GoodCount,
			TotalCount: allPrev[i].TotalCount,
		})
	}

	return allChecks
}

func getCheckStatus(w http.ResponseWriter, r *http.Request) {
	checkID := mux.Vars(r)["check"]

	data, err := primaryDB.GetCheck(checkID, "id")
	if err != nil {
		apiLogger.Errorf("Could not get check by ID, error:\n%s", err.Error())
	}

	status := status{ID: strconv.FormatUint(uint64(data.ID), 10), Title: data.Title, Status: data.Status, GoodCount: data.GoodCount, TotalCount: data.TotalCount}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func getAllChecks(w http.ResponseWriter, r *http.Request) {
	apiLogger.Infoln("Get All Checks")
	allChecks := getAllStatuses()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allChecks)
}
