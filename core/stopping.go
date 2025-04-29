package core

import (
	"github.com/rs/zerolog"
)

type stoppingHandler struct {
	logger      *zerolog.Logger
	coreData    *coreData
	shipControl ShipControl
}

func newStoppingHandler(logger *zerolog.Logger, coreData *coreData,
	shipControl ShipControl) *stoppingHandler {
	return &stoppingHandler{
		logger:      logger,
		coreData:    coreData,
		shipControl: shipControl,
	}
}

func (handler *stoppingHandler) OnEnter() {
	handler.logger.Debug().Msg("OnEnter")

	handler.shipControl.SetSpeed("stop")
	handler.shipControl.SetSteering("straight")
}

func (handler *stoppingHandler) OnExit() {
	handler.logger.Debug().Msg("OnExit")
}

func (handler *stoppingHandler) HandleEvent(event Event) string {
	handler.logger.Debug().Msgf("HandleEvent event=%s", event.String())

	switch event {
	case eventShipDataUpdate:
		if handler.coreData.shipData.Speed == "stop" {
			return "ship stopped"
		}
	}

	return ""
}
