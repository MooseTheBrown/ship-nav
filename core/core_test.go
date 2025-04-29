package core

import (
	"os"
	"testing"
	"time"

	"github.com/moosethebrown/ship-nav/core/model"
	"github.com/rs/zerolog"
)

type mockCoreConfigurer struct{}

func (m *mockCoreConfigurer) Declination() float64 {
	return 0.0
}

func (m *mockCoreConfigurer) UpdateBufSize() int {
	return 100
}

func (m *mockCoreConfigurer) TurningSpeed() string {
	return "fwd40"
}

func (m *mockCoreConfigurer) TurningSteeringLeft() string {
	return "left50"
}

func (m *mockCoreConfigurer) TurningSteeringRight() string {
	return "right40"
}

func (m *mockCoreConfigurer) ApproachSpeed() string {
	return "fwd30"
}

func (m *mockCoreConfigurer) FullSpeed() string {
	return "fwd100"
}

func (m *mockCoreConfigurer) ApproachDistance() float64 {
	return 5.0
}

func (m *mockCoreConfigurer) DistanceInaccuracy() float64 {
	return 0.1
}

func TestCore(t *testing.T) {
	mockShipControl := &mockShipControl{}
	mockCoreConfigurer := &mockCoreConfigurer{}
	logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)

	core := NewCore(mockCoreConfigurer, mockShipControl, &logger)
	go core.Run()
	defer core.Stop()

	position := &model.Position{
		Latitude:  56.412695,
		Longitude: 43.843618,
	}
	core.UpdatePosition(position)

	waypoint := &model.Waypoint{
		Latitude:  56.402099,
		Longitude: 43.859839,
	}
	core.AddWaypoint(waypoint)
	waypoint = &model.Waypoint{
		Latitude:  56.376828,
		Longitude: 43.876562,
	}
	core.AddWaypoint(waypoint)
	homeWaypoint := &model.Waypoint{
		Latitude:  56.412695,
		Longitude: 43.843618,
	}
	core.SetHomeWaypoint(homeWaypoint)

	bearing := model.NewBearing(0.0)
	bearing.SetInt(0, 0)
	core.UpdateBearing(bearing)

	// idle state
	time.Sleep(10 * time.Millisecond)

	if core.fsm.CurrentState() != "idle" {
		t.Errorf("Expected core state to be idle, got %s",
			core.fsm.CurrentState())
	}

	// turning to the first waypoint
	core.StartNavigation()
	time.Sleep(10 * time.Millisecond)

	if core.fsm.CurrentState() != "turning" {
		t.Errorf("Expected core state to be turning, got %s",
			core.fsm.CurrentState())
	}
	if mockShipControl.speed != "fwd40" {
		t.Errorf("Expected turning speed fwd40, got %s", mockShipControl.speed)
	}
	if mockShipControl.steering != "right40" {
		t.Errorf("Expected steering to be right40, got %s", mockShipControl.steering)
	}

	// moving to the first waypoint
	bearing = model.NewBearing(0.0)
	bearing.SetFloat((56.402099 - 56.412695), (43.859839 - 43.843618))
	core.UpdateBearing(bearing)
	time.Sleep(10 * time.Millisecond)

	if core.fsm.CurrentState() != "moving" {
		t.Errorf("Expected core state to be moving, got %s",
			core.fsm.CurrentState())
	}
	if mockShipControl.speed != "fwd100" {
		t.Errorf("Expected speed to be fwd100, got %s",
			mockShipControl.speed)
	}
	if mockShipControl.steering != "straight" {
		t.Errorf("Expected steering to be straight, got %s",
			mockShipControl.steering)
	}

	// approaching the first waypoint
	position = &model.Position{
		Latitude:  56.402098,
		Longitude: 43.859838,
	}
	core.UpdatePosition(position)
	time.Sleep(10 * time.Millisecond)

	if core.fsm.CurrentState() != "moving" {
		t.Errorf("Expected core state to be moving, got %s",
			core.fsm.CurrentState())
	}
	if mockShipControl.speed != "fwd30" {
		t.Errorf("Expected speed to be fwd30, got %s", mockShipControl.speed)
	}
	if mockShipControl.steering != "straight" {
		t.Errorf("Expected steering to be straight, got %s",
			mockShipControl.steering)
	}

	// reach the first waypoint, start turning to the second
	position = &model.Position{
		Latitude:  56.402099,
		Longitude: 43.859839,
	}
	core.UpdatePosition(position)
	time.Sleep(10 * time.Millisecond)

	if core.fsm.CurrentState() != "turning" {
		t.Errorf("Expected core state to be turning, got %s",
			core.fsm.CurrentState())
	}
	if mockShipControl.speed != "fwd40" {
		t.Errorf("Expected speed to be fwd40, got %s", mockShipControl.speed)
	}
	if mockShipControl.steering != "right40" {
		t.Errorf("Expected steering to be right40, got %s",
			mockShipControl.steering)
	}

	// net loss
	core.NetworkLost()
	time.Sleep(10 * time.Millisecond)

	if core.fsm.CurrentState() != "turning home" {
		t.Errorf("Expected core state to be turning home, got %s",
			core.fsm.CurrentState())
	}
	if mockShipControl.speed != "fwd40" {
		t.Errorf("Expected speed to be fwd40, got %s", mockShipControl.speed)
	}
	if mockShipControl.steering != "right40" {
		t.Errorf("Expected steering to be right40, got %s",
			mockShipControl.steering)
	}

	// complete turn to the home waypoint, start moving home
	bearing = model.NewBearing(0.0)
	bearing.SetFloat((56.412695 - 56.402099), (43.843618 - 43.859839))
	core.UpdateBearing(bearing)
	time.Sleep(10 * time.Millisecond)

	if core.fsm.CurrentState() != "moving home" {
		t.Errorf("Expected core state to be moving home, got %s",
			core.fsm.CurrentState())
	}
	if mockShipControl.speed != "fwd100" {
		t.Errorf("Expected speed to be fwd100, got %s", mockShipControl.speed)
	}
	if mockShipControl.steering != "straight" {
		t.Errorf("Expected steering to be straight, got %s",
			mockShipControl.steering)
	}

	// approach home
	position = &model.Position{
		Latitude:  56.412665,
		Longitude: 43.843612,
	}
	core.UpdatePosition(position)
	time.Sleep(10 * time.Millisecond)

	if core.fsm.CurrentState() != "moving home" {
		t.Errorf("Expected core state to be moving home, got %s",
			core.fsm.CurrentState())
	}
	if mockShipControl.speed != "fwd30" {
		t.Errorf("Expected speed to be fwd30, got %s", mockShipControl.speed)
	}
	if mockShipControl.steering != "straight" {
		t.Errorf("Expected steering to be straight, got %s",
			mockShipControl.steering)
	}

	// reach home
	position = &model.Position{
		Latitude:  56.412695,
		Longitude: 43.843618,
	}
	core.UpdatePosition(position)
	time.Sleep(10 * time.Millisecond)

	if core.fsm.CurrentState() != "stopping" {
		t.Errorf("Expected core state to be stopping, got %s",
			core.fsm.CurrentState())
	}
	if mockShipControl.speed != "stop" {
		t.Errorf("Expected speed to be stop, got %s", mockShipControl.speed)
	}
	if mockShipControl.steering != "straight" {
		t.Errorf("Expected steering to be straight, got %s",
			mockShipControl.steering)
	}

	// move to idle state when the ship is stopped
	shipData := &model.ShipData{
		Speed:    "stop",
		Steering: "straight",
	}
	core.UpdateShipData(shipData)
	time.Sleep(10 * time.Millisecond)
	if core.fsm.CurrentState() != "idle" {
		t.Errorf("Expected core state to be idle, got %s",
			core.fsm.CurrentState())
	}
	if mockShipControl.speed != "stop" {
		t.Errorf("Expected speed to be stop, got %s", mockShipControl.speed)
	}
	if mockShipControl.steering != "straight" {
		t.Errorf("Expected steering to be straight, got %s",
			mockShipControl.steering)
	}
}
