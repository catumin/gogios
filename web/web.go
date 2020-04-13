package web

import (
	"net/http"
	"os"
	"strconv"
	"text/template"

	"github.com/bkasin/gogios"
	"github.com/bkasin/gogios/helpers/config"
	"github.com/google/logger"
)

var LayoutDir string = "/usr/share/gogios/views"
var refresh = 3
var data = []gogios.Check{}
var bootstrap *template.Template
var webLogger *logger.Logger

type ViewData struct {
	Checks  []gogios.Check
	Refresh int
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
	tmpl, err := bootstrap.ParseFiles(LayoutDir + "/checks.html")
	if err != nil {
		panic(err)
	}

	// Inject data into template
	vd := ViewData{
		Checks:  data,
		Refresh: refresh * 60,
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

	refresh = int(conf.Options.Interval.Duration.Minutes())
	data, err = conf.Databases[0].Database.GetAllChecks()
	if err != nil {
		webLogger.Errorf("Failed to read rows from database. Error:\n%s", err.Error())
	}
	webLogger.Infof("Refresh rate: %d", refresh)

	// Serve static files while preventing directory listing
	fs := http.FileServer(http.Dir(LayoutDir))
	bootstrap, err = template.ParseFiles(LayoutDir + "/checks.html")
	if err != nil {
		panic(err)
	}

	http.Handle("/", fs)
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
