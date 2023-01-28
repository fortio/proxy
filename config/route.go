package config

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"fortio.org/fortio/log"
)

type JSONURL struct {
	Str string
	URL url.URL
}

// Route configuration. Does Host/Prefix match to destination, see Match* functions.
// (only host,port,scheme part of Destination URL are used).
type Route struct {
	// Host or * or empty to match any host (* without a Prefix must be the last rule)
	Host string
	// Prefix or empty for any
	Prefix string
	// Destination url string.
	Destination JSONURL
}

// UnmarshalJSON is needed to get a URL from json
// until golang does it on its own.
func (j *JSONURL) UnmarshalJSON(b []byte) error {
	l := len(b)
	if l == 0 {
		return nil
	}
	l--
	if l == 0 || b[0] != '"' || b[l] != '"' {
		return fmt.Errorf("invalid url string %q", b)
	}
	j.Str = "Proxy to " + string(b[1:l])
	return j.URL.UnmarshalBinary(b[1:l])
}

// Match checks if there is a match.
// Path prefix has to match (or be empty)
// Host has to match or route spec be "*" (last entry).

func (r *Route) MatchServerReq(req *http.Request) bool {
	// Server requests have host:port in req.Host (and nothing host/port related in req.URL)
	return r.MatchHostAndPath(req.Host, req.URL.Path)
}

func (r *Route) MatchHostAndPath(hostPort, path string) bool {
	host := hostPort
	idx := strings.LastIndex(hostPort, ":")
	if idx != -1 && hostPort[len(hostPort)-1] != ']' { // could be [:some:ip:v6:addr] without :port
		host = hostPort[:idx]
	}
	log.LogVf("path is %q. req host is %q -> %q", path, hostPort, host)
	if r.Prefix != "" && !strings.HasPrefix(path, r.Prefix) {
		return false
	}
	if r.Host == "" || r.Host == "*" {
		return true
	}
	return r.Host == host
}
