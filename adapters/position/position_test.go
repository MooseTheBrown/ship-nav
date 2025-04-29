package position

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"os"
	"testing"
	"time"

	"github.com/moosethebrown/ship-nav/core/model"
	"github.com/rs/zerolog"
)

const (
	testSocket = "/tmp/position_testsock"
)

type mockPositionUpdater struct {
	position *model.Position
}

func (m *mockPositionUpdater) UpdatePosition(position *model.Position) {
	m.position = position
}

type mockBearingUpdater struct {
	bearing *model.Bearing
}

func (m *mockBearingUpdater) UpdateBearing(bearing *model.Bearing) {
	m.bearing = bearing
}

type mockInfoProvider struct {
	socket      string
	listener    net.Listener
	connections []net.Conn
	calibration bool
}

func newMockInfoProvider(socket string) *mockInfoProvider {
	return &mockInfoProvider{
		socket:      socket,
		connections: make([]net.Conn, 0),
	}
}

func (p *mockInfoProvider) run() {
	var err error
	p.listener, err = net.Listen("unix", p.socket)
	if err != nil {
		fmt.Printf("Error opening socket %s for listening: %s\n", p.socket, err.Error())
		return
	}
	defer p.listener.Close()

	for {
		conn, err := p.listener.Accept()
		if err == nil {
			p.connections = append(p.connections, conn)
			go p.handleConn(conn)
		} else {
			fmt.Printf("Error accepting socket connection: %s\n", err.Error())
			return
		}
	}
}

func (p *mockInfoProvider) stop() {
	if p.listener != nil {
		p.listener.Close()
	}

	for _, conn := range p.connections {
		if conn != nil {
			conn.Close()
		}
	}
}

func (p *mockInfoProvider) handleConn(conn net.Conn) {
	defer conn.Close()

	for {
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("Failed to read data from unix socket: %s\n", err.Error())
			return
		}

		rq := &IPCRequest{}
		err = json.Unmarshal(buf[:n], &rq)
		if err != nil {
			fmt.Printf("Failed to unmarshal request: %s\n", err.Error())
			return
		}

		var respData []byte
		switch rq.Cmd {
		case CmdGetGPS:
			resp := GPSInfoResponse{
				NumSatellites: 11,
				Latitude:      56.363358,
				Longitude:     43.902243,
				SpeedKnots:    3.0,
				SpeedKm:       5.56,
			}
			respData, err = json.Marshal(resp)
			if err != nil {
				fmt.Printf("Failed to marshal response: %s\n", err.Error())
			}
		case CmdGetMagnetometer:
			resp := MagnetometerInfoResponse{
				X: 13281,
				Y: 11824,
				Z: -9584,
			}
			respData, err = json.Marshal(resp)
			if err != nil {
				fmt.Printf("Failed to marshal response: %s\n", err.Error())
			}
		case CmdStartCalibration:
			p.calibration = true
			fallthrough
		case CmdStopCalibration:
			p.calibration = false
			resp := CalibrationResponse{Success: true}
			respData, err = json.Marshal(resp)
			if err != nil {
				fmt.Printf("Failed to marshal response: %s\n", err.Error())
			}
		}

		if respData != nil {
			_, err = conn.Write(respData)
			if err != nil {
				fmt.Printf("Failed to send response: %s\n", err.Error())
				return
			}
		}
	}
}

type mockConfigurer struct{}

func (c *mockConfigurer) PositionSocketName() string {
	return testSocket
}

func (c *mockConfigurer) PositionPollingInterval() int64 {
	return 500
}

func (c *mockConfigurer) Declination() float64 {
	return 0.0
}

func setupTest() (*mockInfoProvider, *mockBearingUpdater, *mockPositionUpdater, *Adapter) {
	mockInfoProvider := newMockInfoProvider(testSocket)
	mockBearingUpdater := &mockBearingUpdater{}
	mockPositionUpdater := &mockPositionUpdater{}

	logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)
	adapter := NewAdapter(&logger, &mockConfigurer{}, mockPositionUpdater, mockBearingUpdater)

	return mockInfoProvider, mockBearingUpdater, mockPositionUpdater, adapter
}

func TestDataUpdate(t *testing.T) {
	mockInfoProvider, mockBearingUpdater, mockPositionUpdater, adapter := setupTest()

	go mockInfoProvider.run()
	defer mockInfoProvider.stop()
	time.Sleep(50 * time.Millisecond)

	go adapter.Run()
	defer adapter.Stop()

	time.Sleep(550 * time.Millisecond)
	// position and bearing updates should have happened

	if mockBearingUpdater.bearing == nil {
		t.Fatal("expected bearing to be updated")
	}
	expectedAngle := math.Atan2(float64(11824), float64(13281))
	if mockBearingUpdater.bearing.Angle() != expectedAngle {
		t.Errorf("expected bearing angle to be %f, got %f",
			expectedAngle, mockBearingUpdater.bearing.Angle())
	}

	if mockPositionUpdater.position == nil {
		t.Fatal("expected position to be updated")
	}
	if mockPositionUpdater.position.NumSatellites != 11 {
		t.Errorf("expected number of satellites to be 11, got %d",
			mockPositionUpdater.position.NumSatellites)
	}
	if mockPositionUpdater.position.Latitude != 56.363358 {
		t.Errorf("expected latitude to be 56.363358, got %f",
			mockPositionUpdater.position.Latitude)
	}
	if mockPositionUpdater.position.Longitude != 43.902243 {
		t.Errorf("expected longitude to be 43.902243, got %f",
			mockPositionUpdater.position.Longitude)
	}
	if mockPositionUpdater.position.SpeedKnots != 3.0 {
		t.Errorf("expected speed in knots to be 3.0, got %f",
			mockPositionUpdater.position.SpeedKnots)
	}
	if mockPositionUpdater.position.SpeedKm != 5.56 {
		t.Errorf("expected speed in km to be 5.56, got %f",
			mockPositionUpdater.position.SpeedKm)
	}
}
