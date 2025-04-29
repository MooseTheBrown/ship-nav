package core

import (
	"math"
	"testing"

	"github.com/moosethebrown/ship-nav/core/model"
	"github.com/rs/zerolog"
)

type mockShipControl struct {
	speed    string
	steering string
}

func (m *mockShipControl) SetSpeed(speed string) {
	m.speed = speed
}

func (m *mockShipControl) SetSteering(steering string) {
	m.steering = steering
}

func TestTurningOnEnter(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	waypoints := model.NewWaypoints()
	waypoints.AddWaypoint(&model.Waypoint{
		Latitude:  56.33956,
		Longitude: 43.98449,
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

	handler := &turningHandler{
		logger:               &logger,
		coreData:             coreData,
		shipControl:          shipControl,
		turningSpeed:         "fwd30",
		turningSteeringLeft:  "left40",
		turningSteeringRight: "right40",
	}

	handler.OnEnter()

	tolerance := 0.0001
	if math.Abs(coreData.targetBearing.AngleDeg()-(-92.665815)) > tolerance {
		t.Errorf("Expected target bearing to be -92.665815, got %f", coreData.targetBearing.AngleDeg())
	}

	// with current bearing = 0, target bearing -92.665815, ship is expected to turn left
	if shipControl.steering != "left40" {
		t.Errorf("Expected steering to be left40, got %s", shipControl.steering)
	}

	if shipControl.speed != "fwd30" {
		t.Errorf("Expected speed to be fwd30, got %s", shipControl.speed)
	}
}

func TestTurningOnExit(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	coreData := &coreData{
		curBearing:    model.NewBearing(0.0),
		targetBearing: model.NewBearing(0.0),
	}

	handler := &turningHandler{
		logger:   &logger,
		coreData: coreData,
	}

	coreData.targetBearing.SetInt(170, 3)

	handler.OnExit()

	if coreData.targetBearing.AngleDeg() != 0 {
		t.Errorf("Expected target bearing to be 0, got %f", coreData.targetBearing.AngleDeg())
	}
}

func TestTurningEventBearingUpdate(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	waypoints := model.NewWaypoints()
	waypoints.AddWaypoint(&model.Waypoint{
		Latitude:  56.33956,
		Longitude: 43.98449,
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

	handler := &turningHandler{
		logger:               &logger,
		coreData:             coreData,
		shipControl:          shipControl,
		turningSpeed:         "fwd30",
		turningSteeringLeft:  "left40",
		turningSteeringRight: "right40",
	}

	handler.OnEnter()
	// target bearing is -92.665815 degrees here

	event := Event(eventBearingUpdate)
	coreData.curBearing.SetInt(2, -1) // angle = -26.56 degrees
	transition := handler.HandleEvent(event)
	if transition != "" {
		t.Errorf("Expected empty transition, got %s", transition)
	}

	coreData.curBearing.SetFloat(-0.04655, -1) // angle = -92.66519457519553 degrees
	transition = handler.HandleEvent(event)
	if transition != "bearing adjust" {
		t.Errorf("Expected bearing adjust transition, got %s", transition)
	}
}

func TestTurningEventPositionUpdate(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	waypoints := model.NewWaypoints()
	waypoints.AddWaypoint(&model.Waypoint{
		Latitude:  56.33956,
		Longitude: 43.98449,
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

	handler := &turningHandler{
		logger:               &logger,
		coreData:             coreData,
		shipControl:          shipControl,
		turningSpeed:         "fwd30",
		turningSteeringLeft:  "left40",
		turningSteeringRight: "right40",
	}

	handler.OnEnter()
	// target bearing is -92.665815 degrees here

	coreData.position.Latitude = 56.33938
	coreData.position.Longitude = 43.99413
	event := Event(eventPositionUpdate)

	transition := handler.HandleEvent(event)
	if transition != "" {
		t.Errorf("Expected empty transition, got %s", transition)
	}
	tolerance := 0.0001
	if math.Abs(coreData.targetBearing.AngleDeg()-(-88.93028610071595)) > tolerance {
		t.Errorf("Expected target bearing to be -88.93028610071595, got %f", coreData.targetBearing.AngleDeg())
	}
}

func TestTurningEventNetLoss(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	coreData := &coreData{
		curBearing:    model.NewBearing(0.0),
		targetBearing: model.NewBearing(0.0),
	}

	handler := &turningHandler{
		logger:   &logger,
		coreData: coreData,
	}

	transition := handler.HandleEvent(Event(eventNetLoss))
	if transition != "net loss stop" {
		t.Errorf("Expected net loss stop transition, got %s", transition)
	}

	coreData.homeWaypoint = &model.Waypoint{
		Latitude:  56.33938,
		Longitude: 43.99413,
	}

	transition = handler.HandleEvent(Event(eventNetLoss))
	if transition != "net loss home" {
		t.Errorf("Expected net loss home transition, got %s", transition)
	}
}

func TestTurningEventNavStop(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	coreData := &coreData{
		curBearing:    model.NewBearing(0.0),
		targetBearing: model.NewBearing(0.0),
	}

	handler := &turningHandler{
		logger:   &logger,
		coreData: coreData,
	}

	transition := handler.HandleEvent(Event(eventNavStop))
	if transition != "nav stop" {
		t.Errorf("Expected nav stop home transition, got %s", transition)
	}
}

func TestTurningEventWaypointsSet(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	waypoints := model.NewWaypoints()
	waypoints.AddWaypoint(&model.Waypoint{
		Latitude:  56.33956,
		Longitude: 43.98449,
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

	handler := &turningHandler{
		logger:               &logger,
		coreData:             coreData,
		shipControl:          shipControl,
		turningSpeed:         "fwd30",
		turningSteeringLeft:  "left40",
		turningSteeringRight: "right40",
	}

	handler.OnEnter()
	// target bearing is -92.665815 degrees here, ship turns left

	coreData.waypoints = model.NewWaypoints()
	coreData.waypoints.AddWaypoint(&model.Waypoint{
		Latitude:  56.338651,
		Longitude: 44.000639,
	})
	transition := handler.HandleEvent(Event(eventWaypointsSet))
	if transition != "" {
		t.Errorf("Expected empty transition, got %s", transition)
	}
	if shipControl.steering != "right40" {
		t.Errorf("Expected steering to be right40, got %s", shipControl.steering)
	}
}

func TestTurningEventWaypointsCleared(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	waypoints := model.NewWaypoints()
	waypoints.AddWaypoint(&model.Waypoint{
		Latitude:  56.33956,
		Longitude: 43.98449,
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

	handler := &turningHandler{
		logger:               &logger,
		coreData:             coreData,
		shipControl:          shipControl,
		turningSpeed:         "fwd30",
		turningSteeringLeft:  "left40",
		turningSteeringRight: "right40",
	}

	handler.OnEnter()

	coreData.waypoints = model.NewWaypoints()
	transition := handler.HandleEvent(Event(eventWaypointsCleared))
	if transition != "waypoints cleared" {
		t.Errorf("Expected waypoints cleared transition, got %s", transition)
	}
}
