package config

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"fortio.org/fortio/log"
)

type JSONURL struct {
	URL url.URL
}

// Route configuration. for now only does Host match to destination, see Director function.
// (only host,port,scheme part of Destination URL are used)
type Route struct {
	// Host or * to match any host (* must be the last rule)
	Host string
	// Port or 0/not specified for any port
	Port uint
	// Destination url string.
	Destination JSONURL
}

func (j *JSONURL) UnmarshalJSON(b []byte) error {
	l := len(b)
	if l == 0 {
		return nil
	}
	l--
	if l == 0 || b[0] != '"' || b[l] != '"' {
		return fmt.Errorf("invalid url string %q", b)
	}
	return j.URL.UnmarshalBinary(b[1:l])
}

// Match checks if there is a match.
func (r *Route) Match(req *http.Request) bool {
	portStr := req.URL.Port()
	hostFromReq := req.URL.Host
	idx := strings.LastIndex(hostFromReq, ":")
	if idx != -1 && hostFromReq[len(hostFromReq)-1] != ']' {
		p := hostFromReq[idx+1:]
		if p != portStr && portStr != "" {
			log.Warnf("unexoected port missmatch %q vs %q", p, portStr)
		}
		portStr = p
		hostFromReq = hostFromReq[:idx]
		log.Infof("changing host to %q", hostFromReq)
	}
	var port uint
	if portStr != "" {
		p, err := strconv.Atoi(portStr)
		if err != nil {
			log.Errf("Unexpected unable to convert %q to port: %v", portStr, err)
		}
		port = uint(p)
	} else {
		port = 80
		if req.URL.Scheme == "https" {
			port = 443
		}
	}
	log.Infof("port is %q scheme %s -> %d - req url host is %q", portStr, req.URL.Scheme, port, req.URL.Host)
	if r.Port != 0 && port != r.Port {
		return false
	}
	if r.Host == "*" {
		return true
	}
	if r.Host == hostFromReq {
		return true
	}
	return false
}
