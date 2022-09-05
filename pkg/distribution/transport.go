package distribution

import "net/http"

type Transport struct {
	roundTripper http.RoundTripper
	funcs        func(*http.Request)
}

// RoundTrip handles each http request
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.funcs(req)
	return t.roundTripper.RoundTrip(req)
}

// NewTransport creates a new Transport
func NewTransport(funcs func(*http.Request)) http.RoundTripper {
	var tran = &Transport{
		roundTripper: http.DefaultTransport,
		funcs:        funcs,
	}
	return tran
}
