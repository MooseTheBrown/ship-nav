package core

import (
	"github.com/rs/zerolog"
)

type movingHandler struct {
	logger             *zerolog.Logger
	coreData           *coreData
	shipControl        ShipControl
	approachSpeed      string
	fullSpeed          string
	approachDistance   float64
	distanceInaccuracy float64
}

func newMovingHandler(logger *zerolog.Logger, coreData *coreData, shipControl ShipControl,
	approachSpeed string, fullSpeed string, approachDistance float64,
	distanceInaccuracy float64) *movingHandler {
	return &movingHandler{
		logger:             logger,
		coreData:           coreData,
		shipControl:        shipControl,
		approachSpeed:      approachSpeed,
		fullSpeed:          fullSpeed,
		approachDistance:   approachDistance,
		distanceInaccuracy: distanceInaccuracy,
	}
}

func (handler *movingHandler) OnEnter() {
	handler.logger.Debug().Msg("OnEnter")

	handler.shipControl.SetSteering("straight")

	distance := handler.coreData.position.DistanceMeters(handler.coreData.waypoints.GetNextWaypoint())
	handler.logger.Debug().Msgf("distance to target = %f", distance)

	handler.setSpeed(distance)
}

func (handler *movingHandler) OnExit() {
	handler.logger.Debug().Msg("OnExit")
}

func (handler *movingHandler) HandleEvent(event Event) string {
	handler.logger.Debug().Msgf("HandleEvent event=%s", event.String())

	switch event {
	case eventPositionUpdate:
		distance := handler.coreData.position.DistanceMeters(handler.coreData.waypoints.GetNextWaypoint())
		handler.logger.Debug().Msgf("distance to target = %f", distance)
		if distance <= handler.distanceInaccuracy {
			// waypoint reached
			handler.coreData.waypoints.WaypointReached()
			if handler.coreData.waypoints.GetNextWaypoint() == nil {
				return "last waypoint"
			} else {
				return "waypoint"
			}
		} else {
			// continue moving
			handler.setSpeed(distance)
		}
	case eventNetLoss:
		if handler.coreData.homeWaypoint != nil {
			handler.logger.Info().Msg("net loss home")
			return "net loss home"
		} else {
			return "net loss stop"
		}
	case eventNavStop:
		return "nav stop"
	case eventWaypointsSet:
		return "waypoints set"
	case eventWaypointsCleared:
		return "waypoints cleared"
	}

	return ""
}

func (handler *movingHandler) setSpeed(distance float64) {
	if distance < handler.approachDistance {
		handler.shipControl.SetSpeed(handler.approachSpeed)
	} else {
		handler.shipControl.SetSpeed(handler.fullSpeed)
	}
}
