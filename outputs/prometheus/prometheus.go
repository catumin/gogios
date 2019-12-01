package prometheus

import (
	"net/http"

	"github.com/bkasin/gogios/helpers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Prometheus defines the HTTP server to host /metrics on
func Prometheus(conf helpers.Config) {
	// create a new mux server
	server := http.NewServeMux()
	// register a new handler for the /metrics endpoint
	server.Handle("/metrics", promhttp.Handler())
	// start an http server using the mux server
	http.ListenAndServe(conf.Prometheus.IP+":"+conf.Prometheus.Port, server)
}
