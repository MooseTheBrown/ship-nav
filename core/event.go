package core

const (
	eventUndefined = iota
	eventPositionUpdate
	eventHomeWaypointUpdate
	eventBearingUpdate
	eventShipDataUpdate
	eventWaypointsSet
	eventWaypointAdded
	eventWaypointsCleared
	eventNavStart
	eventNavStop
	eventNetLoss
)

type Event uint16

func (e Event) String() string {
	switch e {
	case eventPositionUpdate:
		return "eventPositionUpdate"
	case eventHomeWaypointUpdate:
		return "eventHomeWaypointUpdate"
	case eventBearingUpdate:
		return "eventBearingUpdate"
	case eventShipDataUpdate:
		return "eventShipDataUpdate"
	case eventWaypointsSet:
		return "eventWaypointsSet"
	case eventWaypointAdded:
		return "eventWaypointAdded"
	case eventWaypointsCleared:
		return "eventWaypointsCleared"
	case eventNavStart:
		return "eventNavStart"
	case eventNavStop:
		return "eventNavStop"
	case eventNetLoss:
		return "eventNetLoss"
	default:
		return "undefined"
	}
}
