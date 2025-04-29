package core

import (
	"testing"

	"github.com/moosethebrown/ship-nav/core/model"
	"github.com/rs/zerolog"
)

func TestMovingHomeOnEnter(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	coreData := &coreData{
		position: &model.Position{
			Latitude:  56.34000,
			Longitude: 43.99394,
		},
		curBearing:    model.NewBearing(0.0),
		targetBearing: model.NewBearing(0.0),
		homeWaypoint: &model.Waypoint{
			Latitude:  56.333284,
			Longitude: 44.008402,
		},
	}

	shipControl := &mockShipControl{}

	handler := newMovingHomeHandler(&logger, coreData, shipControl, "fwd50", "fwd100", 50.0, 0.5)
	handler.OnEnter()

	if shipControl.speed != "fwd100" {
		t.Errorf("Expected speed to be fwd100, got %s", shipControl.speed)
	}
	if shipControl.steering != "straight" {
		t.Errorf("Expected steering to be straight, got %s", shipControl.steering)
	}
}

func TestMovingHomeEventPositionUpdate(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	coreData := &coreData{
		position: &model.Position{
			Latitude:  56.34000,
			Longitude: 43.99394,
		},
		curBearing:    model.NewBearing(0.0),
		targetBearing: model.NewBearing(0.0),
		homeWaypoint: &model.Waypoint{
			Latitude:  56.333284,
			Longitude: 44.008402,
		},
	}

	shipControl := &mockShipControl{}

	handler := newMovingHomeHandler(&logger, coreData, shipControl, "fwd50", "fwd100", 50.0, 6)
	handler.OnEnter()

	// approach home position
	coreData.position.Latitude = 56.333588
	coreData.position.Longitude = 44.008166
	transition := handler.HandleEvent(Event(eventPositionUpdate))
	if transition != "" {
		t.Errorf("Expected empty transition, got %s", transition)
	}
	if shipControl.speed != "fwd50" {
		t.Errorf("Expected speed to be fwd50, got %s", shipControl.speed)
	}

	// reach home position
	coreData.position.Latitude = 56.333328
	coreData.position.Longitude = 44.008408
	transition = handler.HandleEvent(Event(eventPositionUpdate))
	if transition != "home reached" {
		t.Errorf("Expected home reached transition, got %s", transition)
	}
}

func TestMovingHomeEventNavStop(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	coreData := &coreData{
		position: &model.Position{
			Latitude:  56.34000,
			Longitude: 43.99394,
		},
		curBearing:    model.NewBearing(0.0),
		targetBearing: model.NewBearing(0.0),
		homeWaypoint: &model.Waypoint{
			Latitude:  56.333284,
			Longitude: 44.008402,
		},
	}

	shipControl := &mockShipControl{}

	handler := newMovingHomeHandler(&logger, coreData, shipControl, "fwd50", "fwd100", 50.0, 6)
	handler.OnEnter()

	transition := handler.HandleEvent(Event(eventNavStop))
	if transition != "nav stop" {
		t.Errorf("Expected nav stop transition, got %s", transition)
	}
}
