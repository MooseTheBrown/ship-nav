package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/moosethebrown/ship-nav/adapters/network"
	"github.com/moosethebrown/ship-nav/adapters/position"
	"github.com/moosethebrown/ship-nav/adapters/ship"
	"github.com/moosethebrown/ship-nav/config"
	"github.com/moosethebrown/ship-nav/core"
	"github.com/rs/zerolog"
)

type App struct {
	conf            *config.Config
	logger          *zerolog.Logger
	theCore         *core.Core
	shipAdapter     *ship.Adapter
	positionAdapter *position.Adapter
	networkAdapter  *network.Adapter
	wg              sync.WaitGroup
}

func NewApp(conf *config.Config) *App {
	logLevel, err := zerolog.ParseLevel(conf.LogLevel)
	if err != nil {
		fmt.Printf("Invalid logLevel: %s, error: %s", conf.LogLevel, err.Error())
		logLevel = zerolog.InfoLevel
	}

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger().Level(logLevel)

	app := &App{
		conf:   conf,
		logger: &logger,
	}

	app.init()

	return app
}

func (app *App) Start() {
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		app.shipAdapter.Run()
	}()

	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		app.positionAdapter.Run()
	}()

	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		app.networkAdapter.Run()
	}()

	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		app.theCore.Run()
	}()
}

func (app *App) Stop() {
	app.networkAdapter.Stop()
	app.positionAdapter.Stop()
	app.shipAdapter.Stop()
	app.theCore.Stop()
	app.wg.Wait()
}

func (app *App) init() {
	shipAdapterLogger := app.logger.With().Str("component", "ship-adapter").Logger()
	app.shipAdapter = ship.NewAdapter(&shipAdapterLogger, app.conf, nil)

	coreLogger := app.logger.With().Str("component", "core").Logger()
	app.theCore = core.NewCore(app.conf, app.shipAdapter, &coreLogger)

	app.shipAdapter.SetShipDataUpdater(app.theCore)

	positionAdapterLogger := app.logger.With().Str("component", "position-adapter").Logger()
	app.positionAdapter = position.NewAdapter(&positionAdapterLogger, app.conf, app.theCore, app.theCore)

	networkAdapterLogger := app.logger.With().Str("component", "network-adapter").Logger()
	app.networkAdapter = network.NewAdapter(app.conf.NetworkSocketName(), app.theCore, app.theCore,
		app.theCore, app.theCore, app.theCore, &networkAdapterLogger)
}
