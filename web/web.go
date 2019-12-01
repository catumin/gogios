package web

import (
	"net/http"
	"strconv"
	"text/template"

	"github.com/bkasin/gogios/helpers"
	"github.com/gorilla/mux"
)

// Have the page refresh default to 180. Gets set in ServePage
var refresh = 3

// httpsRedirect redirects HTTP requests to HTTPS
func httpsRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(
		w, r,
		"https://"+r.Host+r.URL.String(),
		http.StatusMovedPermanently,
	)
}

// renderTemplate renders page after passing some data to the HTML template
func renderChecks(w http.ResponseWriter, r *http.Request) {
	// Load template from disk
	tmpl := template.Must(template.ParseFiles("/opt/gingertechengine/checks.html"))
	// Inject data into template
	data := refresh * 60
	helpers.Log.Println("Checks page accessed")
	tmpl.Execute(w, data)
}

// ServePage hosts a server based on options from the config file
func ServePage(conf helpers.Config) {
	refresh = conf.Options.Interval
	// Serve static files while preventing directory listing
	fs := http.FileServer(http.Dir("/opt/gingertechengine/"))
	router := mux.NewRouter().StrictSlash(true)
	router.Handle("/", fs)
	router.HandleFunc("/checks", renderChecks)

	if conf.WebOptions.ExposeAPI {
		// All API calls will live under /api/
		router.HandleFunc("/api/", apiHome)
	}

	if conf.WebOptions.SSL {
		go http.ListenAndServeTLS(conf.WebOptions.IP+":"+strconv.Itoa(conf.WebOptions.HTTPSPort), conf.WebOptions.TLSCert, conf.WebOptions.TLSKey, router)
	}

	if conf.WebOptions.Redirect {
		go http.ListenAndServe(conf.WebOptions.IP+":"+strconv.Itoa(conf.WebOptions.HTTPPort), http.HandlerFunc(httpsRedirect))
	} else {
		go http.ListenAndServe(conf.WebOptions.IP+":"+strconv.Itoa(conf.WebOptions.HTTPPort), router)
	}
}
