package main

import (
	"flag"
	"runtime/debug"

	"fortio.org/fortio/dflag"
	"fortio.org/fortio/dflag/configmap"
	"fortio.org/fortio/log"
	"fortio.org/proxy/config"
)

func main() {
	configs := dflag.DynJSON(flag.CommandLine, "routes.json", &[]config.Route{}, "json list of `routes`")
	fullVersion := flag.Bool("version", false, "Show full version info")
	configDir := flag.String("config", "",
		"Config directory `path` to watch for changes of dynamic flags (empty for no watch)")
	flag.Parse()
	version := "unknown"
	binfo, ok := debug.ReadBuildInfo()
	if ok {
		version = binfo.Main.Version
	}
	log.Infof("Fortio Proxy %s starting", version)
	if *fullVersion {
		log.Infof("Buildinfo: %s", binfo.String())
	}
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
