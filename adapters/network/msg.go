package network

const (
	rqTypeQuery = "query"
	rqTypeCmd   = "cmd"
)

const (
	cmdNavStart        = "nav_start"
	cmdNavStop         = "nav_stop"
	cmdNetLoss         = "net_loss"
	cmdSetWaypoints    = "set_waypoints"
	cmdAddWaypoint     = "add_waypoint"
	cmdClearWaypoints  = "clear_waypoints"
	cmdSetHomeWaypoint = "set_home_waypoint"
)

type Waypoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Request struct {
	Type      string      `json:"type"`
	Cmd       string      `json:"cmd"`
	Waypoints []*Waypoint `json:"waypoints"`
}

type PositionData struct {
	NumSatellites int8    `json:"num_satellites"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	SpeedKnots    float64 `json:"speed_knots"`
	SpeedKm       float64 `json:"speed_km"`
	Angle         float64 `json:"angle"`
}

type ShipData struct {
	Speed    string `json:"speed"`
	Steering string `json:"steering"`
}

type QueryResponse struct {
	PositionData *PositionData `json:"positionData"`
	ShipData     *ShipData     `json:"shipData"`
	Waypoints    []*Waypoint   `json:"waypoints"`
	Error        string        `json:"error"`
}

type CommandResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}
