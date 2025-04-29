package core

import (
	"testing"

	"github.com/moosethebrown/ship-nav/core/model"
	"github.com/rs/zerolog"
)

func TestStoppingOnEnter(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	coreData := &coreData{
		curBearing:    model.NewBearing(0.0),
		targetBearing: model.NewBearing(0.0),
	}

	shipControl := &mockShipControl{}

	handler := newStoppingHandler(&logger, coreData, shipControl)
	handler.OnEnter()

	if shipControl.speed != "stop" {
		t.Errorf("Expected speed to be stop, got %s", shipControl.speed)
	}
	if shipControl.steering != "straight" {
		t.Errorf("Expected steering to be straight, got %s", shipControl.steering)
	}
}

func TestStoppingEventShipDataUpdate(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	coreData := &coreData{
		curBearing:    model.NewBearing(0.0),
		targetBearing: model.NewBearing(0.0),
		shipData:      &model.ShipData{},
	}

	shipControl := &mockShipControl{}

	handler := newStoppingHandler(&logger, coreData, shipControl)
	handler.OnEnter()

	coreData.shipData.Speed = "fwd10"
	transition := handler.HandleEvent(Event(eventShipDataUpdate))
	if transition != "" {
		t.Errorf("Expected empty transition, got %s", transition)
	}

	coreData.shipData.Speed = "stop"
	transition = handler.HandleEvent(Event(eventShipDataUpdate))
	if transition != "ship stopped" {
		t.Errorf("Expected ship stopped transition, got %s", transition)
	}
}
