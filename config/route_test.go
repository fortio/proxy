package config

import (
	"encoding/json"
	"net/http"
	"testing"

	"fortio.org/fortio/log"
)

func TestMatch(t *testing.T) {
	// Do 2 tests at once, deserialize json and use "Destination" as a match against the rule itself:
	tests := []struct {
		jsonRoute string
		match     bool
	}{
		{`{"host": "*", "destination": "http://www.google.com/"}`, true},                                  // 0
		{`{"host": "*", "port": 25, "destination": "http://www.google.com/"}`, false},                     // 1
		{`{"host": "*", "port": 80, "destination": "http://www.google.com/"}`, true},                      // 2
		{`{"host": "www.google.com", "port": 80, "destination": "hTTp://www.google.com/?foo=bar"}`, true}, // 3
		{`{"host": "*", "port": 80, "destination": "http://www.google.com:80/"}`, true},                   // 4
		{`{"host": "*", "port": 443, "destination": "http://www.google.com/"}`, false},                    // 5
		{`{"host": "*", "port": 443, "destination": "https://www.google.com/"}`, true},                    // 6
		{`{"host": "*", "port": 443, "destination": "https://www.google.com:443/"}`, true},                // 7
		{`{"host": "www.google.com", "port": 443, "destination": "https://www.google.com/"}`, true},       // 8
		{`{"host": "www.google.com", "port": 443, "destination": "https://www.google.com:443/"}`, true},   // 9
		{`{"host": "www.google.com", "port": 443, "destination": "http://www.google.com:443/"}`, true},    // 10
		{`{"host": "www.google.com", "port": 443, "destination": "https://google.com/"}`, false},          // 11
		{`{"host": "[:1:2", "port": 80, "destination": "http://[:1:2:443]/"}`, false},                     // 12
		{`{"host": "[:1:2:443]", "port": 443, "destination": "https://[:1:2:443]/"}`, true},               // 13
		{`{"host": "[:1:2:443]", "port": 80, "destination": "http://[:1:2:443]/"}`, true},                 // 14
		{`{"host": "[:1:2:443]", "port": 443, "destination": "https://[:1:2:443]/"}`, true},               // 15
		{`{"host": "[:1:2:443]", "destination": "https://[:1:2:443]:673/"}`, true},                        // 16
		{`{"host": "[:1:2:443]", "port": 443, "destination": "https://[:1:2:443]:673/"}`, false},          // 17
		{`{"host": "[:1:2:443]", "port": 673, "destination": "https://[:1:2:443]:673/"}`, true},           // 18
		{`{"host": "[:1:2:443]", "port": 80, "destination": "https://[:1:2:443]:80/"}`, true},             // 19
	}
	for i, tst := range tests {
		route := Route{}
		err := json.Unmarshal([]byte(tst.jsonRoute), &route)
		if err != nil {
			t.Errorf("unmarshal error: %v", err)
		}
		if route.Destination.URL.String() == "" || route.Host == "" {
			t.Errorf("Unexpected empty deserialization: %s -> %+v", tst.jsonRoute, route)
		}
		urlStr := route.Destination.URL.String()
		req, err := http.NewRequest(http.MethodGet, urlStr, nil)
		if err != nil {
			t.Errorf("Unable to deserialize %q %+v: %v", urlStr, req, err)
		}
		res := route.Match(req)
		log.Infof("%d expecting %v got %v for H %s P %d D %s vs Req scheme %s host %s port %s", i, tst.match, res, route.Host, route.Port,
			route.Destination.URL.String(), req.URL.Scheme, req.URL.Host, req.URL.Port())
		if res != tst.match {
			t.Errorf("Mismatch %d expected %v for H %s P %d D %s vs Req scheme %s host %s port %s", i, tst.match, route.Host, route.Port,
				route.Destination.URL.String(), req.URL.Scheme, req.URL.Host, req.URL.Port())
		}
	}
}
