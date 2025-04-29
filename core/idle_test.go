package core

import (
	"testing"

	"github.com/moosethebrown/ship-nav/core/model"
	"github.com/rs/zerolog"
)

func TestIdleEventNavStart(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	coreData := &coreData{
		curBearing:    model.NewBearing(0.0),
		targetBearing: model.NewBearing(0.0),
	}

	handler := newIdleHandler(&logger, coreData)
	handler.OnEnter()

	transition := handler.HandleEvent(Event(eventNavStart))
	if transition != "nav start" {
		t.Errorf("Expected nav start transition, got %s", transition)
	}
}

func TestIdleEventNetLoss(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	coreData := &coreData{
		curBearing:    model.NewBearing(0.0),
		targetBearing: model.NewBearing(0.0),
		homeWaypoint: &model.Waypoint{
			Latitude:  56.333284,
			Longitude: 44.008402,
		},
	}

	handler := newIdleHandler(&logger, coreData)
	handler.OnEnter()

	transition := handler.HandleEvent(Event(eventNetLoss))
	if transition != "net loss home" {
		t.Errorf("Expected net loss home transition, got %s", transition)
	}
}
