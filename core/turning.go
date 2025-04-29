package core

import (
	"github.com/moosethebrown/ship-nav/core/model"
	"github.com/rs/zerolog"
)

type turningHandler struct {
	logger               *zerolog.Logger
	coreData             *coreData
	shipControl          ShipControl
	turningSpeed         string
	turningSteeringLeft  string
	turningSteeringRight string
}

func newTurningHandler(logger *zerolog.Logger, coreData *coreData, shipControl ShipControl,
	turningSpeed string, turningSteeringLeft string, turningSteeringRight string) *turningHandler {
	return &turningHandler{
		logger:               logger,
		coreData:             coreData,
		shipControl:          shipControl,
		turningSpeed:         turningSpeed,
		turningSteeringLeft:  turningSteeringLeft,
		turningSteeringRight: turningSteeringRight,
	}
}

func (handler *turningHandler) OnEnter() {
	handler.logger.Debug().Msg("OnEnter")

	handler.calculateTargetBearing(handler.coreData.waypoints.GetNextWaypoint())
	handler.steerToTarget()
	handler.shipControl.SetSpeed(handler.turningSpeed)
}

func (handler *turningHandler) OnExit() {
	handler.logger.Debug().Msg("OnExit")

	handler.coreData.targetBearing.SetInt(0, 0)
}

func (handler *turningHandler) HandleEvent(event Event) string {
	handler.logger.Debug().Msgf("HandleEvent event=%s", event.String())

	switch event {
	case eventBearingUpdate:
		deltaAngle := handler.coreData.targetBearing.AngleDeg() -
			handler.coreData.curBearing.AngleDeg()
		handler.logger.Debug().Msgf("delta angle = %f", deltaAngle)
		if ((deltaAngle < 0.1) && (deltaAngle >= 0)) || ((deltaAngle > -0.1) && (deltaAngle <= 0)) {
			// close enough
			handler.logger.Info().Msgf("delta angle = %f, turning is completed", deltaAngle)
			return "bearing adjust"
		}
	case eventPositionUpdate:
		handler.calculateTargetBearing(handler.coreData.waypoints.GetNextWaypoint())
		handler.steerToTarget()
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
		handler.calculateTargetBearing(handler.coreData.waypoints.GetNextWaypoint())
		handler.steerToTarget()
	case eventWaypointsCleared:
		return "waypoints cleared"
	}

	return ""
}

func (handler *turningHandler) calculateTargetBearing(waypoint *model.Waypoint) {
	diffLat := waypoint.Latitude - handler.coreData.position.Latitude
	diffLong := waypoint.Longitude - handler.coreData.position.Longitude
	handler.coreData.targetBearing.SetFloat(diffLat, diffLong)
	handler.logger.Debug().Msgf("current bearing = %f", handler.coreData.curBearing.AngleDeg())
	handler.logger.Debug().Msgf("target bearing = %f", handler.coreData.targetBearing.AngleDeg())
}

func (handler *turningHandler) steerToTarget() {
	deltaAngle := handler.coreData.targetBearing.AngleDeg() -
		handler.coreData.curBearing.AngleDeg()
	if (deltaAngle > 180) || ((deltaAngle > -180) && (deltaAngle < 0)) {
		handler.shipControl.SetSteering(handler.turningSteeringLeft)
	} else {
		handler.shipControl.SetSteering(handler.turningSteeringRight)
	}
}
