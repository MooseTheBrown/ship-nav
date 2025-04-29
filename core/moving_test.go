package core

import (
	"testing"

	"github.com/moosethebrown/ship-nav/core/model"
	"github.com/rs/zerolog"
)

func TestMovingOnEnter(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	waypoints := model.NewWaypoints()
	waypoints.AddWaypoint(&model.Waypoint{
		Latitude:  56.33956,
		Longitude: 43.98449,
	})
	waypoints.AddWaypoint(&model.Waypoint{
		Latitude:  56.333015,
		Longitude: 44.007853,
	})

	coreData := &coreData{
		position: &model.Position{
			Latitude:  56.34000,
			Longitude: 43.99394,
		},
		curBearing:    model.NewBearing(0.0),
		targetBearing: model.NewBearing(0.0),
		waypoints:     waypoints,
	}

	shipControl := &mockShipControl{}

	handler := newMovingHandler(&logger, coreData, shipControl, "fwd30", "fwd80", 50.0, 0.5)
	handler.OnEnter()

	if shipControl.speed != "fwd80" {
		t.Errorf("Expected speed to be fwd80, got %s", shipControl.speed)
	}
	if shipControl.steering != "straight" {
		t.Errorf("Expected steering to be straight, got %s", shipControl.steering)
	}
}

func TestMovingEventPositionUpdate(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	waypoints := model.NewWaypoints()
	waypoints.AddWaypoint(&model.Waypoint{
		Latitude:  56.33956,
		Longitude: 43.98449,
	})
	waypoints.AddWaypoint(&model.Waypoint{
		Latitude:  56.333015,
		Longitude: 44.007853,
	})

	coreData := &coreData{
		position: &model.Position{
			Latitude:  56.34,
			Longitude: 43.99394,
		},
		curBearing:    model.NewBearing(0.0),
		targetBearing: model.NewBearing(0.0),
		waypoints:     waypoints,
	}

	shipControl := &mockShipControl{}

	handler := newMovingHandler(&logger, coreData, shipControl, "fwd30", "fwd80", 50.0, 0.5)
	handler.OnEnter()

	// approach first waypoint
	coreData.position.Latitude = 56.339582
	coreData.position.Longitude = 43.984714
	transition := handler.HandleEvent(Event(eventPositionUpdate))
	if transition != "" {
		t.Errorf("Expected empty transition, got %s", transition)
	}
	if shipControl.speed != "fwd30" {
		t.Errorf("Expected speed to be fwd30, got %s", shipControl.speed)
	}

	// move to the first waypoint close enough to consider it reached
	coreData.position.Latitude = 56.339557
	coreData.position.Longitude = 43.984488
	transition = handler.HandleEvent(Event(eventPositionUpdate))
	if transition != "waypoint" {
		t.Errorf("Expected waypoint transition, got %s", transition)
	}
	if coreData.waypoints.GetNextWaypoint().Latitude != 56.333015 {
		t.Errorf("Expected next waypoint latitude to be 56.333015, got %f",
			coreData.waypoints.GetNextWaypoint().Latitude)
	}
	if coreData.waypoints.GetNextWaypoint().Longitude != 44.007853 {
		t.Errorf("Expected next waypoint latitude to be 44.007853, got %f",
			coreData.waypoints.GetNextWaypoint().Latitude)
	}

	// move to the last waypoint
	coreData.position.Latitude = 56.333015
	coreData.position.Longitude = 44.007853
	transition = handler.HandleEvent(Event(eventPositionUpdate))
	if transition != "last waypoint" {
		t.Errorf("Expected last waypoint transition, got %s", transition)
	}
	if coreData.waypoints.GetNextWaypoint() != nil {
		t.Errorf("Expected nil next waypoint, got %f, %f",
			coreData.waypoints.GetNextWaypoint().Latitude,
			coreData.waypoints.GetNextWaypoint().Longitude)
	}
}

func TestMovingEventNetLoss(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	waypoints := model.NewWaypoints()
	waypoints.AddWaypoint(&model.Waypoint{
		Latitude:  56.33956,
		Longitude: 43.98449,
	})
	waypoints.AddWaypoint(&model.Waypoint{
		Latitude:  56.333015,
		Longitude: 44.007853,
	})

	coreData := &coreData{
		position: &model.Position{
			Latitude:  56.34,
			Longitude: 43.99394,
		},
		curBearing:    model.NewBearing(0.0),
		targetBearing: model.NewBearing(0.0),
		waypoints:     waypoints,
	}

	shipControl := &mockShipControl{}

	handler := newMovingHandler(&logger, coreData, shipControl, "fwd30", "fwd80", 50.0, 0.5)
	handler.OnEnter()

	transition := handler.HandleEvent(Event(eventNetLoss))
	if transition != "net loss stop" {
		t.Errorf("Expected net loss stop transition, got %s", transition)
	}

	coreData.homeWaypoint = &model.Waypoint{
		Latitude:  56.333015,
		Longitude: 44.007853,
	}

	transition = handler.HandleEvent(Event(eventNetLoss))
	if transition != "net loss home" {
		t.Errorf("Expected net loss home transition, got %s", transition)
	}
}

func TestMovingEventNavStop(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	waypoints := model.NewWaypoints()
	waypoints.AddWaypoint(&model.Waypoint{
		Latitude:  56.33956,
		Longitude: 43.98449,
	})
	waypoints.AddWaypoint(&model.Waypoint{
		Latitude:  56.333015,
		Longitude: 44.007853,
	})

	coreData := &coreData{
		position: &model.Position{
			Latitude:  56.34,
			Longitude: 43.99394,
		},
		curBearing:    model.NewBearing(0.0),
		targetBearing: model.NewBearing(0.0),
		waypoints:     waypoints,
	}

	shipControl := &mockShipControl{}

	handler := newMovingHandler(&logger, coreData, shipControl, "fwd30", "fwd80", 50.0, 0.5)
	handler.OnEnter()

	transition := handler.HandleEvent(Event(eventNavStop))
	if transition != "nav stop" {
		t.Errorf("Expected nav stop transition, got %s", transition)
	}
}

func TestMovingEventWaypointsSet(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	waypoints := model.NewWaypoints()
	waypoints.AddWaypoint(&model.Waypoint{
		Latitude:  56.33956,
		Longitude: 43.98449,
	})

	coreData := &coreData{
		position: &model.Position{
			Latitude:  56.34,
			Longitude: 43.99394,
		},
		curBearing:    model.NewBearing(0.0),
		targetBearing: model.NewBearing(0.0),
		waypoints:     waypoints,
	}

	shipControl := &mockShipControl{}

	handler := newMovingHandler(&logger, coreData, shipControl, "fwd30", "fwd80", 50.0, 0.5)
	handler.OnEnter()

	coreData.waypoints = model.NewWaypoints()
	coreData.waypoints.AddWaypoint(&model.Waypoint{
		Latitude:  56.333015,
		Longitude: 44.007853,
	})
	transition := handler.HandleEvent(Event(eventWaypointsSet))
	if transition != "waypoints set" {
		t.Errorf("Expected waypoints set transition, got %s", transition)
	}
}

func TestMovingEventWaypointsCleared(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	waypoints := model.NewWaypoints()
	waypoints.AddWaypoint(&model.Waypoint{
		Latitude:  56.33956,
		Longitude: 43.98449,
	})

	coreData := &coreData{
		position: &model.Position{
			Latitude:  56.34,
			Longitude: 43.99394,
		},
		curBearing:    model.NewBearing(0.0),
		targetBearing: model.NewBearing(0.0),
		waypoints:     waypoints,
	}

	shipControl := &mockShipControl{}

	handler := newMovingHandler(&logger, coreData, shipControl, "fwd30", "fwd80", 50.0, 0.5)
	handler.OnEnter()

	coreData.waypoints = model.NewWaypoints()
	transition := handler.HandleEvent(Event(eventWaypointsCleared))
	if transition != "waypoints cleared" {
		t.Errorf("Expected waypoints cleared transition, got %s", transition)
	}
}
