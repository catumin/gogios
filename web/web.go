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

var LayoutDir string = "/usr/share/gogios/views"
var title string
var navbar string
var logo string
var refresh = 3
var data = []gogios.Check{}
var webLogger *logger.Logger
var primaryDB gogios.Database

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

// httpsRedirect redirects HTTP requests to HTTPS
func httpsRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(
		w, r,
		"https://"+r.Host+r.URL.String(),
		http.StatusMovedPermanently,
	)
}

// renderChecks renders page after passing some data to the HTML template
func renderChecks(w http.ResponseWriter, r *http.Request) {
	// Load template from disk
	tmpl, err := template.ParseFiles(LayoutDir + "/checks.html")
	if err != nil {
		panic(err)
	}
	table := genTable()

	// Inject data into template
	vd := ViewData{
		Checks:  table,
		Refresh: refresh * 60,
		Title:   title,
		NavBar:  navbar,
	}

	tmpl.Execute(w, vd)
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(LayoutDir + "/index.html")
	if err != nil {
		panic(err)
	}
	table := genTable()

	// Inject data into template
	vd := ViewData{
		Checks: table,
		Title:  title,
		NavBar: navbar,
	}

	tmpl.Execute(w, vd)
}

// ServePage hosts a server based on options from the config file
func ServePage(conf *config.Config) {
	var err error

	// Prepare the web logger
	log, err := os.OpenFile("/var/log/gogios/web.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}
	defer log.Close()

	webLogger = logger.Init("WebLog", conf.Options.Verbose, true, log)
	defer webLogger.Close()

	primaryDB = conf.Databases[0].Database
	refresh = int(conf.Options.Interval.Duration.Minutes())
	title = conf.WebOptions.Title
	navbar = conf.WebOptions.NavBar
	logo = conf.WebOptions.Logo

	data, err = conf.Databases[0].Database.GetAllChecks()
	if err != nil {
		webLogger.Errorf("Failed to read rows from database. Error:\n%s", err.Error())
	}
	webLogger.Infof("Refresh rate: %d", refresh)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(LayoutDir+"/static"))))
	http.HandleFunc("/", mainPage)
	http.HandleFunc("/checks", renderChecks)

	if conf.WebOptions.SSL {
		go http.ListenAndServeTLS(conf.WebOptions.IP+":"+strconv.Itoa(conf.WebOptions.HTTPSPort), conf.WebOptions.TLSCert, conf.WebOptions.TLSKey, nil)
	}

	if conf.WebOptions.Redirect {
		http.ListenAndServe(conf.WebOptions.IP+":"+strconv.Itoa(conf.WebOptions.HTTPPort), http.HandlerFunc(httpsRedirect))
	} else {
		http.ListenAndServe(conf.WebOptions.IP+":"+strconv.Itoa(conf.WebOptions.HTTPPort), nil)
	}
}

// UpdateWebData - Update the table each time new data is available
func UpdateWebData(conf *config.Config) {
	var err error

	data, err = conf.Databases[0].Database.GetAllChecks()
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
