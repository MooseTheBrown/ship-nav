package model

import "testing"

func TestDistanceMeters(t *testing.T) {
	pos := &Position{
		Latitude:  56.326773,
		Longitude: 44.006053,
	}

	waypoint := &Waypoint{
		Latitude:  56.318266,
		Longitude: 44.015766,
	}

	distance := pos.DistanceMeters(waypoint)
	tolerance := 1.0
	if (distance - 1120) > tolerance {
		t.Errorf("Expected distance to be 1120 meters, got %f", distance)
	}

	waypoint.Latitude = 56.326773
	waypoint.Longitude = 44.006053
	distance = pos.DistanceMeters(waypoint)
	if distance != 0 {
		t.Errorf("Expected distance to be 0, got %f", distance)
	}
}
