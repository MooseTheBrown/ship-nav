package core

import (
	"github.com/rs/zerolog"
)

type movingHomeHandler struct {
	movingHandler *movingHandler
}

func newMovingHomeHandler(logger *zerolog.Logger, coreData *coreData,
	shipControl ShipControl, approachSpeed string, fullSpeed string,
	approachDistance float64, distanceInaccuracy float64) *movingHomeHandler {

	movingHandler := newMovingHandler(logger, coreData, shipControl, approachSpeed,
		fullSpeed, approachDistance, distanceInaccuracy)
	return &movingHomeHandler{
		movingHandler: movingHandler,
	}
}

func (handler *movingHomeHandler) OnEnter() {
	handler.movingHandler.logger.Debug().Msg("OnEnter")
	if handler.movingHandler.coreData.homeWaypoint == nil {
		handler.movingHandler.logger.Error().Msg("home waypoint is nil")
		return
	}

	handler.movingHandler.shipControl.SetSteering("straight")

	distance := handler.movingHandler.coreData.position.DistanceMeters(handler.movingHandler.coreData.homeWaypoint)
	handler.movingHandler.setSpeed(distance)
}

func (handler *movingHomeHandler) OnExit() {
	handler.movingHandler.logger.Debug().Msg("OnExit")
}

func (handler *movingHomeHandler) HandleEvent(event Event) string {
	handler.movingHandler.logger.Debug().Msgf("HandleEvent event=%s", event.String())

	switch event {
	case eventPositionUpdate:
		distance := handler.movingHandler.coreData.position.DistanceMeters(handler.movingHandler.coreData.homeWaypoint)
		handler.movingHandler.logger.Debug().Msgf("distance to target = %f", distance)
		if distance <= handler.movingHandler.distanceInaccuracy {
			return "home reached"
		} else {
			handler.movingHandler.setSpeed(distance)
		}
	case eventNavStop:
		return "nav stop"
	}

	return ""
}
