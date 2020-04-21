package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/bkasin/gogios"
	"github.com/bkasin/gogios/helpers/config"
	"github.com/bkasin/gogios/users"
)

func createNewUser(w http.ResponseWriter, r *http.Request) {
	var user gogios.User
	var statusCode int
	var resp map[string]interface{}

	apiLogger.Infoln("Attempting to create new user")

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 104876))
	if err != nil {
		apiLogger.Errorf("Error reading input:\n%v", err.Error())
	}
	if err := r.Body.Close(); err != nil {
		apiLogger.Errorf("Error closing input:\n%v", err.Error())
	}

	if err := json.Unmarshal(body, &user); err != nil {
		apiLogger.Errorf("Create new user JSON error:\n%v", err.Error())
	}

	err = users.CreateUser(user, config.Conf)
	if err != nil {
		apiLogger.Errorf("Create new user error:\n%v", err.Error())
		resp = map[string]interface{}{"status": "Failed", "error": err.Error()}
		statusCode = 422 // Entry could not be processed
	} else {
		resp = map[string]interface{}{"status": "Created", "name": user.Name, "username": user.Username, "password": "Hidden"}
		statusCode = 201 // Status created
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		fmt.Println(err.Error())
		apiLogger.Errorf("Error when sending response about user creation:\n%v", err.Error())
	}
}

func apiLogin(w http.ResponseWriter, r *http.Request) {
	var user gogios.User
	var statusCode int
	var resp map[string]interface{}

	apiLogger.Infoln("Authentication test")

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 104876))
	if err != nil {
		apiLogger.Errorf("Error reading input:\n%v", err.Error())
	}
	if err := r.Body.Close(); err != nil {
		apiLogger.Errorf("Error closing input:\n%v", err.Error())
	}

	if err := json.Unmarshal(body, &user); err != nil {
		statusCode = 422 // Entry could not be processed
		apiLogger.Errorf("Authenitcation test JSON error:\n%v", err.Error())
	}

	resp = users.Login(user.Username, user.Password, primaryDB)

	if resp["status"] == false {
		statusCode = 401 // Unauthorized
		apiLogger.Infof("User %v failed to authenticate", user.Username)
	} else {
		statusCode = 200 // General success
		apiLogger.Infof("User %v successfully authenticated", user.Username)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		fmt.Println(err.Error())
		apiLogger.Errorf("Error when sending response about authentication test:\n%v", err.Error())
	}
}
