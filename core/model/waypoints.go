package model

type Waypoint struct {
	Latitude  float64
	Longitude float64
}

type Waypoints struct {
	waypoints    []*Waypoint
	nextWaypoint int
}

func NewWaypoints() *Waypoints {
	return &Waypoints{
		waypoints: make([]*Waypoint, 0),
	}
}

func (w *Waypoints) SetWaypoints(waypoints []*Waypoint) {
	w.waypoints = waypoints
	w.nextWaypoint = 0
}

func (w *Waypoints) AddWaypoint(waypoint *Waypoint) {
	w.waypoints = append(w.waypoints, waypoint)
}

func (w *Waypoints) GetNextWaypoint() *Waypoint {
	if w.nextWaypoint >= len(w.waypoints) {
		return nil
	} else {
		return w.waypoints[w.nextWaypoint]
	}
}

func (w *Waypoints) WaypointReached() {
	w.nextWaypoint++
}
