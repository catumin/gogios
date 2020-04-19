package web

import (
	"math"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/bkasin/gogios"
	"github.com/bkasin/gogios/helpers/config"
	"github.com/google/logger"
)

var (
	// layoutDir is where all the files for hosting the web interface
	// are stored
	layoutDir = "/usr/share/gogios/views"
	title     string
	navbar    string
	logo      string
	refresh   = 3
	data      = []gogios.Check{}
	webLogger *logger.Logger
	primaryDB gogios.Database
)

type checks struct {
	Title  string
	Status string
	Output string
	Ratio  float64
	Asof   time.Time
}

// ViewData is used to replace variables in the HTML templates
type ViewData struct {
	Checks  []checks
	Refresh int
	Title   string
	NavBar  string
	Logo    string
}

func checksPage(w http.ResponseWriter, r *http.Request) {
	table := genTable()

	// Inject data into template
	vd := ViewData{
		Checks:  table,
		Refresh: refresh * 60,
		Title:   title,
		NavBar:  navbar,
		Logo:    logo,
	}

	render(w, "checks.html", vd, webLogger)
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	table := genTable()

	// Inject data into template
	vd := ViewData{
		Checks: table,
		Title:  title,
		NavBar: navbar,
		Logo:   logo,
	}

	render(w, "index.html", vd, webLogger)
}

// ServePage hosts a server based on options from the config file
func ServePage() {
	var err error

	// Prepare the web logger
	log, err := os.OpenFile("/var/log/gogios/web.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}
	defer log.Close()

	webLogger = logger.Init("WebLog", config.Conf.Options.Verbose, true, log)
	defer webLogger.Close()

	primaryDB = config.Conf.Databases[0].Database
	refresh = int(config.Conf.Options.Interval.Duration.Minutes())
	title = config.Conf.WebOptions.Title
	navbar = config.Conf.WebOptions.NavBar
	logo = config.Conf.WebOptions.Logo

	data, err = config.Conf.Databases[0].Database.GetAllChecks()
	if err != nil {
		webLogger.Errorf("Failed to read rows from database. Error:\n%s", err.Error())
	}
	webLogger.Infof("Refresh rate: %d", refresh)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(layoutDir+"/static"))))
	http.HandleFunc("/", mainPage)
	http.HandleFunc("/checks", checksPage)

	if config.Conf.WebOptions.SSL {
		go http.ListenAndServeTLS(config.Conf.WebOptions.IP+":"+strconv.Itoa(config.Conf.WebOptions.HTTPSPort), config.Conf.WebOptions.TLSCert, config.Conf.WebOptions.TLSKey, nil)
	}

	if config.Conf.WebOptions.Redirect {
		err = http.ListenAndServe(config.Conf.WebOptions.IP+":"+strconv.Itoa(config.Conf.WebOptions.HTTPPort), http.HandlerFunc(httpsRedirect))
		if err != nil {
			webLogger.Errorf("Web server crashed. Error:\n%v", err.Error())
		}
	} else {
		err = http.ListenAndServe(config.Conf.WebOptions.IP+":"+strconv.Itoa(config.Conf.WebOptions.HTTPPort), nil)
		if err != nil {
			webLogger.Errorf("Web server crashed. Error:\n%v", err.Error())
		}
	}
}

// httpsRedirect redirects HTTP requests to HTTPS
func httpsRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(
		w, r,
		"https://"+r.Host+r.URL.String(),
		http.StatusMovedPermanently,
	)
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

// UpdateWebData - Update the table each time new data is available
func UpdateWebData() {
	var err error

	data, err = config.Conf.Databases[0].Database.GetAllChecks()
	if err != nil {
		webLogger.Errorf("Failed to update webpage data from database. Error:\n%s", err.Error())
	}
}

func genTable() []checks {
	var table []checks

	for i := 0; i < len(data); i++ {
		output, err := primaryDB.GetCheckHistory(data[i], 1)
		if err != nil {
			webLogger.Errorf("Error getting history of check:\n%v", err.Error())
		}

		table = append(table, checks{
			Title:  data[i].Title,
			Status: data[i].Status,
			Output: output[0].Output,
			Ratio:  math.Round((float64(data[i].GoodCount) / float64(data[i].TotalCount) * 100)),
			Asof:   data[i].Asof,
		})
	}

	return table
}
