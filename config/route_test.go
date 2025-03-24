package config

import (
	"encoding/json"
	"net/http"
	"testing"

	"fortio.org/log"
)

func TestMatch(t *testing.T) {
	// Do 2 tests at once, deserialize json and use "Destination" as a match against the rule itself:
	tests := []struct {
		jsonRoute string
		match     bool
	}{
		{`{"host": "www.google.com", "destination": "http://www.google.com/"}`, true},                                  // 0
		{`{"host": "*", "prefix": "/foo", "destination": "http://www.google.com/"}`, false},                            // 1
		{`{"host": "*", "prefix": "/foo", "destination": "http://www.google.com/foo/bar"}`, true},                      // 2
		{`{"host": "www.google.com", "prefix": "/test1", "destination": "hTTp://www.google.com/test1?foo=bar"}`, true}, // 3
		{`{"host": "www.google.com", "prefix": "/x/y", "destination": "http://www.google.com:80/x/y"}`, true},          // 4
		{`{"host": "www.google.com", "prefix": "/x/y", "destination": "http://www.google.com/x"}`, false},              // 5
		{`{"host": "www.google.com", "prefix": "/", "destination": "https://www.google.com/"}`, true},                  // 6
		{`{"host": "www.google.com", "prefix": "/ab", "destination": "https://www.google.com:443/abc"}`, true},         // 7
		{`{"host": "www.google.com", "destination": "https://www.google.com/"}`, true},                                 // 8
		{`{"host": "www.google.com", "destination": "https://www.google.com:443/"}`, true},                             // 9
		{`{"host": "www.google.com", "destination": "http://www.google.com:443/"}`, true},                              // 10
		{`{"host": "www.google.com", "destination": "https://google.com/"}`, false},                                    // 11
		{`{"host": "[:1:2", "destination": "http://[:1:2:443]/"}`, false},                                              // 12
		{`{"host": "[:1:2:443]", "destination": "https://[:1:2:443]/"}`, true},                                         // 13
		{`{"host": "[:1:2:443]", "destination": "http://[:1:2:443]/"}`, true},                                          // 14
		{`{"host": "[:1:2:443]", "destination": "https://[:1:2:443]/"}`, true},                                         // 15
		{`{"host": "[:1:2:443]", "destination": "https://[:1:2:443]:673/"}`, true},                                     // 16
		{`{"host": "[:1:2:443]", "prefix": "/x", "destination": "https://[:1:2:443]:673/y"}`, false},                   // 17
		{`{"host": "[:1:2:443]", "prefix": "/x", "destination": "https://[:1:2:443]:673/x"}`, true},                    // 18
		{`{"host": "www.google.com", "destination": "http://www.GOOGLE.com:443/"}`, true},                              // 19
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
		req.Host = req.URL.Host
		res := route.MatchServerReq(req)
		log.Infof("%d expecting %v got %v for H %s P %s D %s vs Req scheme %s host %s port %s",
			i, tst.match, res, route.Host, route.Prefix,
			route.Destination.URL.String(), req.URL.Scheme, req.URL.Host, req.URL.Port())
		if res != tst.match {
			t.Errorf("Mismatch %d expected %v for H %s P %s D %s vs Req scheme %s host %s port %s", i, tst.match, route.Host, route.Prefix,
				route.Destination.URL.String(), req.URL.Scheme, req.URL.Host, req.URL.Port())
		}
	}
}
