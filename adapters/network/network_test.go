package network

import (
	"encoding/json"
	"net"
	"os"
	"testing"
	"time"

	"github.com/moosethebrown/ship-nav/core/model"
	"github.com/rs/zerolog"
)

const (
	testSocket = "/tmp/net_testsock"
)

type mockShipDataProvider struct {
	shipData *model.ShipData
}

func (m *mockShipDataProvider) GetShipData() *model.ShipData {
	return m.shipData
}

type mockPositionDataProvider struct {
	position *model.Position
	bearing  *model.Bearing
}

func (m *mockPositionDataProvider) GetPositionData() (*model.Bearing, *model.Position) {
	return m.bearing, m.position
}

type mockWaypointDataProvider struct {
	waypoints []*model.Waypoint
}

func (m *mockWaypointDataProvider) GetWaypoints() []*model.Waypoint {
	return m.waypoints
}

type mockNavController struct {
	nav     bool
	netLoss bool
}

func (m *mockNavController) StartNavigation() {
	m.nav = true
}

func (m *mockNavController) StopNavigation() {
	m.nav = false
}

func (m *mockNavController) NetworkLost() {
	m.netLoss = true
}

type mockWaypointsUpdater struct {
	waypoints    []*model.Waypoint
	homeWaypoint *model.Waypoint
}

func (m *mockWaypointsUpdater) SetWaypoints(waypoints []*model.Waypoint) {
	m.waypoints = waypoints
}

func (m *mockWaypointsUpdater) AddWaypoint(waypoint *model.Waypoint) {
	m.waypoints = append(m.waypoints, waypoint)
}

func (m *mockWaypointsUpdater) ClearWaypoints() {
	m.waypoints = make([]*model.Waypoint, 0)
}

func (m *mockWaypointsUpdater) SetHomeWaypoint(waypoint *model.Waypoint) {
	m.homeWaypoint = waypoint
}

func TestQuery(t *testing.T) {
	msdp := &mockShipDataProvider{
		shipData: &model.ShipData{
			Speed:    "rev100",
			Steering: "right60",
		},
	}
	mpdp := &mockPositionDataProvider{
		position: &model.Position{
			NumSatellites: 3,
			Latitude:      56.285119,
			Longitude:     44.14972,
			SpeedKnots:    5.24,
			SpeedKm:       9.7,
		},
		bearing: model.NewBearing(0.0),
	}
	mpdp.bearing.SetInt(1, 2)
	mwdp := &mockWaypointDataProvider{}
	mwdp.waypoints = make([]*model.Waypoint, 1)
	mwdp.waypoints[0] = &model.Waypoint{
		Latitude:  56.261437,
		Longitude: 44.191453,
	}
	mnc := &mockNavController{}
	mwu := &mockWaypointsUpdater{}
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger().Level(zerolog.DebugLevel)

	adapter := NewAdapter(testSocket, msdp, mpdp, mwdp, mnc, mwu, &logger)
	go adapter.Run()
	defer adapter.Stop()

	time.Sleep(10 * time.Millisecond)

	conn, err := net.Dial("unix", testSocket)
	if err != nil {
		t.Fatalf("Failed to connect to socket %s: %s",
			testSocket, err.Error())
	}
	defer conn.Close()

	rq := &Request{
		Type: rqTypeQuery,
	}
	rqData, err := json.Marshal(rq)
	if err != nil {
		t.Fatalf("Failed to marshal query: %s", err.Error())
	}

	_, err = conn.Write(rqData)
	if err != nil {
		t.Fatalf("Failed to write to socket: %s", err.Error())
	}

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("Failed to read from socket: %s", err.Error())
	}

	var resp QueryResponse
	err = json.Unmarshal(buf[:n], &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal query response: %s",
			err.Error())
	}

	if resp.PositionData.NumSatellites != 3 {
		t.Errorf("Expected 3 satellites, got %d",
			resp.PositionData.NumSatellites)
	}
	if resp.PositionData.Latitude != 56.285119 {
		t.Errorf("Expected latitude to be 56.285119, got %f",
			resp.PositionData.Latitude)
	}
	if resp.PositionData.Longitude != 44.14972 {
		t.Errorf("Expected longitude to be 44.14972, got %f",
			resp.PositionData.Longitude)
	}
	if resp.PositionData.SpeedKnots != 5.24 {
		t.Errorf("Expected speed to be 5.24 knots, got %f",
			resp.PositionData.SpeedKnots)
	}
	if resp.PositionData.SpeedKm != 9.7 {
		t.Errorf("Expected speed to be 9.7 km/h, got %f",
			resp.PositionData.SpeedKm)
	}
	tolerance := 0.0000000001
	angleDeg := resp.PositionData.Angle
	if (angleDeg - 63.43494882292201) > tolerance {
		t.Errorf("Expected bearing angle to be 63.43494882292201, got %f",
			angleDeg)
	}
	if len(resp.Waypoints) != 1 {
		t.Fatalf("Expected to get 1 waypoint, got %d",
			len(resp.Waypoints))
	}
	if resp.Waypoints[0].Latitude != 56.261437 {
		t.Errorf("Expected waypoint latitude to be 56.261437, got %f",
			resp.Waypoints[0].Latitude)
	}
	if resp.Waypoints[0].Longitude != 44.191453 {
		t.Errorf("Expected waypoint longitude to be 44.191453, got %f",
			resp.Waypoints[0].Longitude)
	}
}

func TestCommand(t *testing.T) {
	msdp := &mockShipDataProvider{}
	mpdp := &mockPositionDataProvider{}
	mwdp := &mockWaypointDataProvider{}
	mnc := &mockNavController{}
	mwu := &mockWaypointsUpdater{}
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger().Level(zerolog.DebugLevel)

	adapter := NewAdapter(testSocket, msdp, mpdp, mwdp, mnc, mwu, &logger)
	go adapter.Run()
	defer adapter.Stop()

	time.Sleep(10 * time.Millisecond)

	conn, err := net.Dial("unix", testSocket)
	if err != nil {
		t.Fatalf("Failed to connect to socket %s: %s",
			testSocket, err.Error())
	}
	defer conn.Close()

	rq := &Request{
		Type: rqTypeCmd,
		Cmd:  cmdNavStart,
	}
	resp, err := sendCommand(conn, rq)
	if err != nil {
		t.Fatalf("Failed to send command: %s", err.Error())
	}
	if resp.Status != "ok" {
		t.Errorf("Expected ok command response status, got %s",
			resp.Status)
	}
	if mnc.nav != true {
		t.Error("Nav status is not set to true")
	}

	rq = &Request{
		Type: rqTypeCmd,
		Cmd:  cmdNavStop,
	}
	resp, err = sendCommand(conn, rq)
	if err != nil {
		t.Fatalf("Failed to send command: %s", err.Error())
	}
	if resp.Status != "ok" {
		t.Errorf("Expected ok command response status, got %s",
			resp.Status)
	}
	if mnc.nav != false {
		t.Error("Nav status is not set to false")
	}

	rq = &Request{
		Type: rqTypeCmd,
		Cmd:  cmdNetLoss,
	}
	resp, err = sendCommand(conn, rq)
	if err != nil {
		t.Fatalf("Failed to send command: %s", err.Error())
	}
	if resp.Status != "ok" {
		t.Errorf("Expected ok command response status, got %s",
			resp.Status)
	}
	if mnc.netLoss != true {
		t.Error("Net loss status is not set to true")
	}

	rq = &Request{
		Type:      rqTypeCmd,
		Cmd:       cmdSetWaypoints,
		Waypoints: make([]*Waypoint, 1),
	}
	rq.Waypoints[0] = &Waypoint{
		Latitude:  56.285119,
		Longitude: 44.14972,
	}
	resp, err = sendCommand(conn, rq)
	if err != nil {
		t.Fatalf("Failed to send command: %s", err.Error())
	}
	if resp.Status != "ok" {
		t.Errorf("Expected ok command response status, got %s",
			resp.Status)
	}
	if len(mwu.waypoints) != 1 {
		t.Fatalf("Expected waypoints length 1, got %d",
			len(mwu.waypoints))
	}
	if mwu.waypoints[0].Latitude != 56.285119 {
		t.Errorf("Expected latitude to be 56.285119, got %f",
			mwu.waypoints[0].Latitude)
	}
	if mwu.waypoints[0].Longitude != 44.14972 {
		t.Errorf("Expected longitude to be 44.14972, got %f",
			mwu.waypoints[0].Longitude)
	}

	rq = &Request{
		Type:      rqTypeCmd,
		Cmd:       cmdAddWaypoint,
		Waypoints: make([]*Waypoint, 1),
	}
	rq.Waypoints[0] = &Waypoint{
		Latitude:  56.261437,
		Longitude: 44.191453,
	}
	resp, err = sendCommand(conn, rq)
	if err != nil {
		t.Fatalf("Failed to send command: %s", err.Error())
	}
	if resp.Status != "ok" {
		t.Errorf("Expected ok command response status, got %s",
			resp.Status)
	}
	if len(mwu.waypoints) != 2 {
		t.Fatalf("Expected waypoints length 2, got %d",
			len(mwu.waypoints))
	}
	if mwu.waypoints[1].Latitude != 56.261437 {
		t.Errorf("Expected wp2 latitude to be 56.261437, got %f",
			mwu.waypoints[1].Latitude)
	}
	if mwu.waypoints[1].Longitude != 44.191453 {
		t.Errorf("Expected wp2 longitude to be 44.191453, got %f",
			mwu.waypoints[1].Longitude)
	}

	rq = &Request{
		Type: rqTypeCmd,
		Cmd:  cmdClearWaypoints,
	}
	resp, err = sendCommand(conn, rq)
	if err != nil {
		t.Fatalf("Failed to send command: %s", err.Error())
	}
	if resp.Status != "ok" {
		t.Errorf("Expected ok command response status, got %s",
			resp.Status)
	}
	if len(mwu.waypoints) != 0 {
		t.Fatalf("Expected waypoints length 0, got %d",
			len(mwu.waypoints))
	}

	rq = &Request{
		Type:      rqTypeCmd,
		Cmd:       cmdSetHomeWaypoint,
		Waypoints: make([]*Waypoint, 1),
	}
	rq.Waypoints[0] = &Waypoint{
		Latitude:  56.261437,
		Longitude: 44.191453,
	}
	resp, err = sendCommand(conn, rq)
	if err != nil {
		t.Fatalf("Failed to send command: %s", err.Error())
	}
	if resp.Status != "ok" {
		t.Errorf("Expected ok command response status, got %s",
			resp.Status)
	}
	if mwu.homeWaypoint == nil {
		t.Fatal("Home waypoint is nil")
	}
	if mwu.homeWaypoint.Latitude != 56.261437 {
		t.Errorf("Expected home waypoint latitude to be 56.261437, got %f",
			mwu.homeWaypoint.Latitude)
	}
	if mwu.homeWaypoint.Longitude != 44.191453 {
		t.Errorf("Expected home waypoint longitude to be 44.191453, got %f",
			mwu.homeWaypoint.Longitude)
	}
}

func sendCommand(conn net.Conn, rq *Request) (*CommandResponse, error) {
	rqData, err := json.Marshal(rq)
	if err != nil {
		return nil, err
	}

	_, err = conn.Write(rqData)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	var resp CommandResponse
	err = json.Unmarshal(buf[:n], &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
