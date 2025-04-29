package model

import "math"

type Position struct {
	NumSatellites int8
	Latitude      float64
	Longitude     float64
	SpeedKnots    float64
	SpeedKm       float64
}

func (p *Position) DistanceMeters(w *Waypoint) float64 {
	if (p == nil) || (w == nil) {
		return 0
	}

	lat1 := p.Latitude * math.Pi / 180
	long1 := p.Longitude * math.Pi / 180
	lat2 := w.Latitude * math.Pi / 180
	long2 := w.Longitude * math.Pi / 180

	return math.Acos(math.Sin(lat1)*math.Sin(lat2)+math.Cos(lat1)*math.Cos(lat2)*math.Cos(long2-long1)) * 6372795
}
