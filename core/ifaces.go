package core

import (
	"github.com/moosethebrown/ship-nav/core/model"
)

// interfaces provided by the core
type PositionUpdater interface {
	UpdatePosition(*model.Position)
}

type BearingUpdater interface {
	UpdateBearing(*model.Bearing)
}

type ShipDataUpdater interface {
	UpdateShipData(*model.ShipData)
}

type WaypointsUpdater interface {
	SetWaypoints([]*model.Waypoint)
	AddWaypoint(*model.Waypoint)
	ClearWaypoints()
	SetHomeWaypoint(*model.Waypoint)
}

type NavigationController interface {
	StartNavigation()
	StopNavigation()
	NetworkLost()
}

type PositionDataProvider interface {
	GetPositionData() (*model.Bearing, *model.Position)
}

type ShipDataProvider interface {
	GetShipData() *model.ShipData
}

type WaypointDataProvider interface {
	GetWaypoints() []*model.Waypoint
}

// interfaces required by the core
type ShipControl interface {
	SetSpeed(string)
	SetSteering(string)
}

type PositionCalibrator interface {
	StartCalibration()
	StopCalibration()
}
