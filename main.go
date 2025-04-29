package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/moosethebrown/ship-nav/config"
)

const (
	defaultConfigFile = "/etc/ship-nav.conf"
)

func main() {
	configFile := flag.String("c", defaultConfigFile, "specify config file location")
	flag.Parse()

	conf, err := config.NewConfig(*configFile)

	if err != nil {
		fmt.Printf("Failed to parse config file: %s", err.Error())
		return
	}

	app := NewApp(conf)

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)

	app.Start()

	<-sigch
	app.Stop()
}
