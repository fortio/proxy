package main

import (
	"flag"

	"fortio.org/fortio/dflag"
	"fortio.org/fortio/dflag/configmap"
	"fortio.org/fortio/log"
	"fortio.org/proxy/config"
)

var (
	version = "dev"
)

func main() {
	configs := dflag.DynJSON(flag.CommandLine, "routes.json", &[]config.Route{}, "json list of `routes`")
	configDir := flag.String("config", "",
		"Config directory `path` to watch for changes of dynamic flags (empty for no watch)")

	log.Infof("Fortio Proxy %s starting", version)
	flag.Parse()
	if *configDir != "" {
		if _, err := configmap.Setup(flag.CommandLine, *configDir); err != nil {
			log.Critf("Unable to watch config/flag changes in %v: %v", *configDir, err)
		}
	}
	routes := configs.Get().(*[]config.Route)
	for i, r := range *routes {
		log.Infof("Route %d: %+v", i, r)
	}
}
