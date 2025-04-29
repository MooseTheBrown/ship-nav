package core

import (
	"math"
	"testing"

	"github.com/moosethebrown/ship-nav/core/model"
	"github.com/rs/zerolog"
)

func TestTurningHomeOnEnter(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	coreData := &coreData{
		position: &model.Position{
			Latitude:  56.34000,
			Longitude: 43.99394,
		},
		homeWaypoint: &model.Waypoint{
			Latitude:  56.33234,
			Longitude: 44.00963,
		},
		curBearing:    model.NewBearing(0.0),
		targetBearing: model.NewBearing(0.0),
	}

	shipControl := &mockShipControl{}

	handler := newTurningHomeHandler(&logger, coreData, shipControl, "fwd20", "left40", "right40")

	handler.OnEnter()

	tolerance := 0.0001
	if math.Abs(coreData.targetBearing.AngleDeg()-(116.022)) > tolerance {
		t.Errorf("Expected target bearing to be 116.022, got %f", coreData.targetBearing.AngleDeg())
	}

	// with current bearing = 0, target bearing 116.022, ship is expected to turn right
	if shipControl.steering != "right40" {
		t.Errorf("Expected steering to be right40, got %s", shipControl.steering)
	}

	if shipControl.speed != "fwd20" {
		t.Errorf("Expected speed to be fwd20, got %s", shipControl.speed)
	}
}

func TestTurningHomeOnExit(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	coreData := &coreData{
		position: &model.Position{
			Latitude:  56.34000,
			Longitude: 43.99394,
		},
		homeWaypoint: &model.Waypoint{
			Latitude:  56.33234,
			Longitude: 44.00963,
		},
		curBearing:    model.NewBearing(0.0),
		targetBearing: model.NewBearing(0.0),
	}

	shipControl := &mockShipControl{}

	handler := newTurningHomeHandler(&logger, coreData, shipControl, "fwd30", "left40", "right40")

	handler.OnExit()

	if coreData.targetBearing.AngleDeg() != 0 {
		t.Errorf("Expected target bearing to be 0, got %f", coreData.targetBearing.AngleDeg())
	}
}

func TestTurningHomeEventBearingUpdate(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	coreData := &coreData{
		position: &model.Position{
			Latitude:  56.34000,
			Longitude: 43.99394,
		},
		homeWaypoint: &model.Waypoint{
			Latitude:  56.33234,
			Longitude: 44.00963,
		},
		curBearing:    model.NewBearing(0.0),
		targetBearing: model.NewBearing(0.0),
	}

	shipControl := &mockShipControl{}

	handler := newTurningHomeHandler(&logger, coreData, shipControl, "fwd30", "left40", "right40")

	handler.OnEnter()
	// target bearing is 116.022 degrees here

	coreData.curBearing.SetInt(-6, 15) // angle = 111.80140948635182 degrees
	transition := handler.HandleEvent(eventBearingUpdate)
	if transition != "" {
		t.Errorf("Expected empty transition, got %s", transition)
	}

	coreData.curBearing.SetFloat(-7.665, 15.7)
	transition = handler.HandleEvent(eventBearingUpdate)
	if transition != "bearing adjust" {
		t.Errorf("Expected bearing adjust transition, got %s", transition)
	}
}

func TestTurningHomeEventPositionUpdate(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	coreData := &coreData{
		position: &model.Position{
			Latitude:  56.34000,
			Longitude: 43.99394,
		},
		homeWaypoint: &model.Waypoint{
			Latitude:  56.33234,
			Longitude: 44.00963,
		},
		curBearing:    model.NewBearing(0.0),
		targetBearing: model.NewBearing(0.0),
	}

	shipControl := &mockShipControl{}

	handler := newTurningHomeHandler(&logger, coreData, shipControl, "fwd30", "left40", "right40")

	handler.OnEnter()
	// target bearing is 116.022 degrees here

	coreData.position.Latitude = 56.33014
	coreData.position.Longitude = 43.98509
	transition := handler.HandleEvent(eventPositionUpdate)
	if transition != "" {
		t.Errorf("Expected empty transition, got %s", transition)
	}
	tolerance := 0.0001
	if math.Abs(coreData.targetBearing.AngleDeg()-(84.87715)) > tolerance {
		t.Errorf("Expected target bearing to be 84.87715, got %f", coreData.targetBearing.AngleDeg())
	}
}

func TestTurningHomeEventNavStop(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	coreData := &coreData{
		position: &model.Position{
			Latitude:  56.34000,
			Longitude: 43.99394,
		},
		homeWaypoint: &model.Waypoint{
			Latitude:  56.33234,
			Longitude: 44.00963,
		},
		curBearing:    model.NewBearing(0.0),
		targetBearing: model.NewBearing(0.0),
	}

	shipControl := &mockShipControl{}

	handler := newTurningHomeHandler(&logger, coreData, shipControl, "fwd30", "left40", "right40")

	handler.OnEnter()

	transition := handler.HandleEvent(eventNavStop)
	if transition != "nav stop" {
		t.Errorf("Expected nav stop transition, got %s", transition)
	}
}
