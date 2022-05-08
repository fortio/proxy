package config

import (
	"fmt"
	"net/url"
)

type JSONURL struct {
	URL url.URL
}

// Route configuration. for now only does Host match to destination, see Director function.
// (only host,port,scheme part of Destination URL are used)
type Route struct {
	Host        string
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
