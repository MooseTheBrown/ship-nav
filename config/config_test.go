package config

import "testing"

func TestConfig(t *testing.T) {
	conf, err := NewConfig("../ship-nav.conf")
	if err != nil {
		t.Fatalf("Failed to parse config file: %s", err.Error())
	}

	if conf.Declination() != 13.62 {
		t.Errorf("Expected declination to be 13.62, got %f", conf.Declination())
	}
	if conf.UpdateBufSize() != 100 {
		t.Errorf("Expected update buf size to be 100, got %d", conf.UpdateBufSize())
	}
	if conf.TurningSpeed() != "fwd30" {
		t.Errorf("Expected turning speed to be fwd30, got %s", conf.TurningSpeed())
	}
	if conf.TurningSteeringLeft() != "left40" {
		t.Errorf("Expected turning steering left to be left40, got %s", conf.TurningSteeringLeft())
	}
	if conf.TurningSteeringRight() != "right40" {
		t.Errorf("Expected turning steering right to be right40, got %s", conf.TurningSteeringRight())
	}
	if conf.ApproachSpeed() != "fwd50" {
		t.Errorf("Expected approach speed to be fwd50, got %s", conf.ApproachSpeed())
	}
	if conf.FullSpeed() != "fwd100" {
		t.Errorf("Expected full speed to be fwd100, got %s", conf.FullSpeed())
	}
	if conf.ApproachDistance() != 10.0 {
		t.Errorf("Expected approach distance to be 10.0, got %f", conf.ApproachDistance())
	}
	if conf.DistanceInaccuracy() != 3.0 {
		t.Errorf("Expected distance inaccuracy to be 3.0, got %f", conf.DistanceInaccuracy())
	}

	if conf.NetworkSocketName() != "/tmp/ship-nav.sock" {
		t.Errorf("Expected network socket name to be /tmp/ship-nav.sock, got %s", conf.NetworkSocketName())
	}

	if conf.PositionSocketName() != "/tmp/ship_position.sock" {
		t.Errorf("Expected position socket name to be /tmp/ship-position.sock, got %s", conf.PositionSocketName())
	}
	if conf.PositionPollingInterval() != 500 {
		t.Errorf("Expected position polling interval to be 500, got %d", conf.PositionPollingInterval())
	}

	if conf.ShipSocketName() != "/tmp/scsocket" {
		t.Errorf("Expected ship socket name to be /tmp/scsocket, got %s", conf.ShipSocketName())
	}
	if conf.ShipPollingInterval() != 500 {
		t.Errorf("Expected ship polling interval to be 500, got %d", conf.ShipPollingInterval())
	}
}
