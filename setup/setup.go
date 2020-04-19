package setup

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/bkasin/gogios/helpers/config"
	"github.com/google/logger"
	"github.com/gorilla/mux"
)

var (
	layoutDir   = "/usr/share/gogios/views"
	setupLogger *logger.Logger
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

// CheckValid that the data entered into the setup form is valid
func (s *Setup) CheckValid() bool {
	// There must be a better way to do this, but I do not currently know it
	s.Errors = make(map[string]string)

	numberPatt := regexp.MustCompile(`\d`)
	yesnoPatt := regexp.MustCompile(`(yes|no)`)
	databasePatt := regexp.MustCompile(`(sqlite|mysql)`)
	ipPatt := regexp.MustCompile(`(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)`)
	ipDomainPatt := regexp.MustCompile(`((?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)|((?:[\w-]+\.)*([\w-]{1,63})(?:\.(?:\w{3}|\w{2}))(?:$|\/)))`)

	if strings.TrimSpace(s.AdminUsername) == "" {
		s.Errors["username"] = "A username is required"
	}
	if strings.TrimSpace(s.AdminPassword) == "" {
		s.Errors["password"] = "A password field is required"
	}

	if !numberPatt.Match([]byte(s.Interval)) {
		s.Errors["interval"] = "This field must be a number"
	}
	if !numberPatt.Match([]byte(s.CheckTimeout)) {
		s.Errors["timeout"] = "This field must be a number"
	}

	if !ipPatt.Match([]byte(s.WebConfig.ListenIP)) {
		s.Errors["webip"] = "A valid IP must be entered"
	}
	if !numberPatt.Match([]byte(s.WebConfig.ListenPort)) {
		s.Errors["webport"] = "This field must be a number"
	}

	if !yesnoPatt.Match([]byte(s.WebConfig.ExposeAPI)) {
		s.Errors["api"] = "It must be specified whether the API will be available"
	} else {
		if s.WebConfig.ExposeAPI == "yes" {
			if !ipPatt.Match([]byte(s.WebConfig.ApiIP)) {
				s.Errors["apiip"] = "A valid IP must be entered"
			}
			if !numberPatt.Match([]byte(s.WebConfig.ApiPort)) {
				s.Errors["apiport"] = "This field must be a number"
			}
		}
	}

	if strings.TrimSpace(s.WebConfig.Title) == "" {
		s.Errors["title"] = "A site title is required"
	}

	if !databasePatt.Match([]byte(s.DatabasePreference)) {
		s.Errors["database"] = "A database must be selected from the dropdown list"
	} else {
		if s.DatabasePreference == "sqlite" {
			if strings.TrimSpace(s.Sqlite.Path) == "" {
				s.Errors["dbpath"] = "A path to a file writeable by gogios is required. The file should not exist yet, but the folders should"
			}
		} else if s.DatabasePreference == "mysql" {
			if !ipDomainPatt.Match([]byte(s.Mysql.Host)) {
				s.Errors["myurl"] = "A valid hostname or IP must be entered"
			}
			if !numberPatt.Match([]byte(s.Mysql.Port)) {
				s.Errors["myport"] = "A valid number must be entered"
			}
			if strings.TrimSpace(s.Mysql.User) == "" {
				s.Errors["myuser"] = "This field must be filled"
			}
			if strings.TrimSpace(s.Mysql.Password) == "" {
				s.Errors["mypass"] = "This field must be filled"
			}
			if strings.TrimSpace(s.Mysql.Database) == "" {
				s.Errors["mydb"] = "This field must be filled and the database must already be created on the server"
			}
		}
	}

	return len(s.Errors) == 0
}

func getForm(w http.ResponseWriter, r *http.Request) {
	render(w, "setup.html", nil, setupLogger)
}

func sendForm(w http.ResponseWriter, r *http.Request) {
	setup := &Setup{
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

	fmt.Println(setup)

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

	printReadyConfig := ""

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

	fmt.Println(printReadyConfig)
}

// FirstSetup displays a webpage where the user is given a brief
// introduction to gogios and configures starting values and the
// first admin user
func FirstSetup() {
	// Prepare the setup logger
	log, err := os.OpenFile("/var/log/gogios/setup.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}
	defer log.Close()

	setupLogger = logger.Init("SetupLog", true, true, log)
	defer setupLogger.Close()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(layoutDir+"/static"))))

	r := mux.NewRouter()
	r.HandleFunc("/", getForm).Methods("GET")
	r.HandleFunc("/", sendForm).Methods("POST")
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
