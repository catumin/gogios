package web

import (
	"fmt"
	"net/http"
	"strconv"
	"text/template"

	"github.com/bkasin/gogios/helpers"
)

// Have the page refresh default to 180. Gets set in ServePage
var refresh = 180

// httpsRedirect redirects HTTP requests to HTTPS
func httpsRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(
		w, r,
		"https://"+r.Host+r.URL.String(),
		http.StatusMovedPermanently,
	)
}

// renderTemplate renders page after passing some data to the HTML template
func renderTemplate(w http.ResponseWriter, r *http.Request) {
	// Load template from disk
	tmpl := template.Must(template.ParseFiles("/opt/gingertechengine/checks.html"))
	// Inject data into template
	data := refresh * 60
	fmt.Println(tmpl)
	tmpl.Execute(w, data)
}

// ServePage hosts a server based on options from the config file
func ServePage(conf helpers.Config) {
	refresh = conf.Options.Interval
	// Serve static files while preventing directory listing
	fs := http.FileServer(http.Dir("/opt/gingertechengine/"))
	http.Handle("/", fs)
	http.HandleFunc("/checks", renderTemplate)

	if conf.WebOptions.Redirect {
		go http.ListenAndServe(":"+strconv.Itoa(conf.WebOptions.HTTPPort), http.HandlerFunc(httpsRedirect))
	} else {
		go http.ListenAndServe(":"+strconv.Itoa(conf.WebOptions.HTTPPort), nil)
	}

	if conf.WebOptions.SSL {
		go http.ListenAndServeTLS(":"+strconv.Itoa(conf.WebOptions.HTTPSPort), conf.WebOptions.TLSCert, conf.WebOptions.TLSKey, nil)
	}
}
