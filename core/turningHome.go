package core

import (
	"github.com/rs/zerolog"
)

type turningHomeHandler struct {
	turningHandler *turningHandler
}

func newTurningHomeHandler(logger *zerolog.Logger, coreData *coreData, shipControl ShipControl,
	turningSpeed string, turningSteeringLeft string, turningSteeringRight string) *turningHomeHandler {
	return &turningHomeHandler{
		turningHandler: &turningHandler{
			logger:               logger,
			coreData:             coreData,
			shipControl:          shipControl,
			turningSpeed:         turningSpeed,
			turningSteeringLeft:  turningSteeringLeft,
			turningSteeringRight: turningSteeringRight,
		},
	}
}

func (handler *turningHomeHandler) OnEnter() {
	handler.turningHandler.logger.Debug().Msg("OnEnter")

	handler.turningHandler.calculateTargetBearing(handler.turningHandler.coreData.homeWaypoint)
	handler.turningHandler.steerToTarget()
	handler.turningHandler.shipControl.SetSpeed(handler.turningHandler.turningSpeed)
}

func (handler *turningHomeHandler) OnExit() {
	handler.turningHandler.logger.Debug().Msg("OnExit")

	handler.turningHandler.coreData.targetBearing.SetInt(0, 0)
}

func (handler *turningHomeHandler) HandleEvent(event Event) string {
	handler.turningHandler.logger.Debug().Msgf("HandleEvent event=%s", event.String())

	switch event {
	case eventNavStop:
		return "nav stop"
	case eventBearingUpdate:
		return handler.turningHandler.HandleEvent(event)
	case eventPositionUpdate:
		handler.turningHandler.calculateTargetBearing(handler.turningHandler.coreData.homeWaypoint)
		return ""
	}

	return ""
}
