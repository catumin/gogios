package setup

import (
	"fmt"
	"net/http"
	"os"
	"text/template"

	"github.com/bkasin/gogios/helpers/config"
	"github.com/google/logger"
	"github.com/gorilla/mux"
)

var (
	layoutDir        = "/usr/share/gogios/views"
	setupLogger      *logger.Logger
	printReadyConfig string
	setup            *Setup
	configPath       string
)

type sqlite struct {
	Path string
}

type mysql struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

type webOptions struct {
	ListenIP   string
	ListenPort string
	ExposeAPI  string
	ApiIP      string
	ApiPort    string
	Title      string
}

// Setup has values that the user enters in the form
// read into it, which then gets used to do first
// setup
type Setup struct {
	AdminName     string
	AdminUsername string
	AdminPassword string

	Interval     string
	CheckTimeout string

	WebConfig webOptions

	DatabasePreference string
	Sqlite             sqlite
	Mysql              mysql

	Errors map[string]string
}

func getForm(w http.ResponseWriter, r *http.Request) {
	render(w, "setup.html", nil, setupLogger)
}

func sendForm(w http.ResponseWriter, r *http.Request) {
	if r.PostFormValue("name") != "" {
		setup = &Setup{
			AdminName:          r.PostFormValue("name"),
			AdminUsername:      r.PostFormValue("username"),
			AdminPassword:      r.PostFormValue("password"),
			Interval:           r.PostFormValue("interval"),
			CheckTimeout:       r.PostFormValue("timeout"),
			DatabasePreference: r.PostFormValue("database"),
		}

		setup.WebConfig = webOptions{
			ListenIP:   r.PostFormValue("webip"),
			ListenPort: r.PostFormValue("webport"),
			ExposeAPI:  r.PostFormValue("api"),
			ApiIP:      r.PostFormValue("apiip"),
			ApiPort:    r.PostFormValue("apiport"),
			Title:      r.PostFormValue("title"),
		}

		setup.Sqlite = sqlite{
			Path: r.PostFormValue("dbpath"),
		}

		setup.Mysql = mysql{
			Host:     r.PostFormValue("myurl"),
			Port:     r.PostFormValue("myport"),
			User:     r.PostFormValue("myuser"),
			Password: r.PostFormValue("mypass"),
			Database: r.PostFormValue("mydb"),
		}

		// Make sure all needed values are valid, make user re-enter
		// if not
		if !setup.CheckValid() {
			render(w, "setup.html", setup, setupLogger)
			return
		}

		if setup.WebConfig.ExposeAPI == "yes" {
			setup.WebConfig.ExposeAPI = "true"
		} else {
			setup.WebConfig.ExposeAPI = "false"
		}

		if setup.DatabasePreference == "sqlite" {
			printReadyConfig = fmt.Sprintf(config.PrintSetupConfig(setup.DatabasePreference),
				setup.Interval, setup.CheckTimeout,
				setup.WebConfig.ListenIP, setup.WebConfig.ListenPort,
				setup.WebConfig.ExposeAPI, setup.WebConfig.ApiIP, setup.WebConfig.ApiPort,
				setup.WebConfig.Title, setup.WebConfig.Title, setup.Sqlite.Path)
		} else if setup.DatabasePreference == "mysql" {
			printReadyConfig = fmt.Sprintf(config.PrintSetupConfig(setup.DatabasePreference),
				setup.Interval, setup.CheckTimeout,
				setup.WebConfig.ListenIP, setup.WebConfig.ListenPort,
				setup.WebConfig.ExposeAPI, setup.WebConfig.ApiIP, setup.WebConfig.ApiPort,
				setup.WebConfig.Title, setup.WebConfig.Title, setup.Mysql.Host, setup.Mysql.Port,
				setup.Mysql.User, setup.Mysql.Password, setup.Mysql.Database)
		}

		http.Redirect(w, r, "/confirm", http.StatusSeeOther)
	} else {
		printReadyConfig = r.PostFormValue("config")
		err := finalizeSetup(printReadyConfig, *setup)
		if err != nil {
			setupLogger.Errorf("Error writing config to base.toml:\n%v", err.Error())
		}
		fmt.Println(setup.DatabasePreference)

		render(w, "setup_done.html", nil, setupLogger)
	}
}

// FirstSetup displays a webpage where the user is given a brief
// introduction to gogios and configures starting values and the
// first admin user
func FirstSetup(configLocation string) {
	// Prepare the setup logger
	log, err := os.OpenFile("/var/log/gogios/setup.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}
	defer log.Close()

	setupLogger = logger.Init("SetupLog", true, true, log)
	defer setupLogger.Close()

	configPath = configLocation

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(layoutDir+"/static"))))

	r := mux.NewRouter()
	r.HandleFunc("/", getForm).Methods("GET")
	r.HandleFunc("/", sendForm).Methods("POST")
	r.HandleFunc("/confirm", confirmConfig).Methods("GET")
	http.Handle("/", r)

	err = http.ListenAndServe(":8411", nil)
	if err != nil {
		setupLogger.Errorf("Web server crashed. Error:\n%v", err.Error())
	}
}

func render(w http.ResponseWriter, filename string, data interface{}, logger *logger.Logger) {
	tmpl, err := template.ParseFiles(layoutDir + "/" + filename)
	if err != nil {
		logger.Errorln(err.Error())
		http.Error(w, "Sorry, something went wrong", http.StatusInternalServerError)
	}

	if err := tmpl.Execute(w, data); err != nil {
		logger.Errorln(err.Error())
		http.Error(w, "Sorry, something went wrong", http.StatusInternalServerError)
	}
}
