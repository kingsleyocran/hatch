package proxy

import (
	"bytes"
	_ "embed"
	"html/template"
	"net/http"
)

//go:embed stopped.html
var stoppedHTML string

var stoppedTmpl = template.Must(template.New("stopped").Parse(stoppedHTML))

type stoppedData struct {
	Domain string
	Port   int
}

func NewStoppedHandler(domain string, port int) http.Handler {
	var buf bytes.Buffer
	stoppedTmpl.Execute(&buf, stoppedData{Domain: domain, Port: port})
	rendered := buf.Bytes()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write(rendered)
	})
}
