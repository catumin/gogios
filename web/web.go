package web

import (
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/bkasin/gogios"
	"github.com/bkasin/gogios/helpers"
	"github.com/bkasin/gogios/helpers/config"
)

var LayoutDir string = "/usr/share/gogios/views"
var refresh = 3
var data = []gogios.Check{}
var bootstrap *template.Template

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
	helpers.Log.Println("Checks page accessed")

	vd := ViewData{
		Checks:  data,
		Refresh: refresh * 60,
	}

	tmpl.Execute(w, vd)
}

// ServePage hosts a server based on options from the config file
func ServePage(conf *config.Config) {
	var err error

	refresh = int(conf.Options.Interval.Duration.Minutes())
	data, err = conf.Databases[0].Database.GetAllRows()
	if err != nil {
		helpers.Log.Println("Failed to read rows from database. Error:")
		helpers.Log.Println(err.Error())
	}

	// Annoying way to refresh the webpage data
	go doEvery(conf.Options.Interval.Duration, updateData, conf)

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
		go http.ListenAndServe(conf.WebOptions.IP+":"+strconv.Itoa(conf.WebOptions.HTTPPort), http.HandlerFunc(httpsRedirect))
	} else {
		go http.ListenAndServe(conf.WebOptions.IP+":"+strconv.Itoa(conf.WebOptions.HTTPPort), nil)
	}
}

func updateData(t time.Time, conf *config.Config) {
	var err error

	data, err = conf.Databases[0].Database.GetAllRows()
	if err != nil {
		helpers.Log.Println("Failed to read rows from database. Error:")
		helpers.Log.Println(err.Error())
	}
}

// doEvery - Run function f every d length of time
func doEvery(d time.Duration, f func(time.Time, *config.Config), conf *config.Config) {
	for x := range time.Tick(d) {
		f(x, conf)
	}
}
