package ship

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/moosethebrown/ship-nav/core/model"
	"github.com/rs/zerolog"
)

const (
	testSocket = "/tmp/ship_testsock"
)

type mockShipConfigurer struct{}

func (m *mockShipConfigurer) ShipSocketName() string {
	return testSocket
}

func (m *mockShipConfigurer) ShipPollingInterval() int64 {
	return 100
}

type mockShipDataUpdater struct {
	shipData *model.ShipData
}

func (m *mockShipDataUpdater) UpdateShipData(shipData *model.ShipData) {
	m.shipData = shipData
}

type mockShipControl struct {
	socketName  string
	listener    net.Listener
	connections []net.Conn
	speed       string
	steering    string
	numQueries  int
	numCmds     int
}

func newMockShipControl(socketName string) *mockShipControl {
	return &mockShipControl{
		socketName:  socketName,
		connections: make([]net.Conn, 0),
	}
}

func (m *mockShipControl) run() {
	var err error
	m.listener, err = net.Listen("unix", m.socketName)
	if err != nil {
		fmt.Printf("Error opening socket %s for listening: %s\n", m.socketName, err.Error())
	}
	defer m.listener.Close()

	for {
		conn, err := m.listener.Accept()
		if err == nil {
			m.connections = append(m.connections, conn)
			go m.handleConn(conn)
		} else {
			fmt.Printf("Error accepting socket connection: %s\n", err.Error())
			return
		}
	}
}

func (m *mockShipControl) stop() {
	if m.listener != nil {
		m.listener.Close()
	}

	for _, conn := range m.connections {
		if conn != nil {
			conn.Close()
		}
	}
}

func (m *mockShipControl) handleConn(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 4096)
	for {
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
		if rq.Type == "cmd" {
			m.numCmds++
			resp := IPCCommandResponse{}
			if rq.Cmd == "set_speed" {
				m.speed = rq.Data
				resp.Status = "ok"
			} else if rq.Cmd == "set_steering" {
				m.steering = rq.Data
				resp.Status = "ok"
			} else {
				resp.Status = "fail"
				resp.Error = fmt.Sprintf("Invalid command %s", rq.Cmd)
			}
			respData, err = json.Marshal(resp)
			if err != nil {
				fmt.Printf("Failed to marshal response: %s\n", err.Error())
			}
		} else if rq.Type == "query" {
			m.numQueries++
			resp := IPCQueryResponse{
				Speed:    m.speed,
				Steering: m.steering,
			}
			respData, err = json.Marshal(resp)
			if err != nil {
				fmt.Printf("Failed to marshal response: %s\n", err.Error())
			}
		} else {
			resp := IPCCommandResponse{
				Status: "fail",
				Error:  "Invalid request type",
			}
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

func setupTest() (*mockShipControl, *mockShipDataUpdater, *Adapter) {
	mockShipControl := newMockShipControl(testSocket)
	mockShipDataUpdater := &mockShipDataUpdater{}
	logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)
	adapter := NewAdapter(&logger, &mockShipConfigurer{}, mockShipDataUpdater)

	return mockShipControl, mockShipDataUpdater, adapter
}

func TestQuery(t *testing.T) {
	mockShipControl, mockShipDataUpdater, adapter := setupTest()

	go mockShipControl.run()
	defer mockShipControl.stop()

	time.Sleep(50 * time.Millisecond)

	go adapter.Run()
	defer adapter.Stop()

	mockShipControl.speed = "rev30"
	mockShipControl.steering = "left10"

	time.Sleep(120 * time.Millisecond)

	if mockShipControl.numQueries != 1 {
		t.Fatalf("Expected to receive 1 query, got %d", mockShipControl.numQueries)
	}
	if mockShipDataUpdater.shipData.Speed != "rev30" {
		t.Errorf("Expected speed to be rev30, got %s", mockShipDataUpdater.shipData.Speed)
	}
	if mockShipDataUpdater.shipData.Steering != "left10" {
		t.Errorf("Expected steering to be left10, got %s", mockShipDataUpdater.shipData.Steering)
	}
}

func TestCommand(t *testing.T) {
	mockShipControl, mockShipDataUpdater, adapter := setupTest()

	go mockShipControl.run()
	defer mockShipControl.stop()

	time.Sleep(50 * time.Millisecond)

	go adapter.Run()
	defer adapter.Stop()

	mockShipControl.speed = "fwd100"
	mockShipControl.steering = "straight"

	adapter.SetSpeed("rev80")
	time.Sleep(120 * time.Millisecond)

	if mockShipControl.numCmds != 1 {
		t.Fatalf("Expected to receive 1 command, got %d", mockShipControl.numCmds)
	}
	if mockShipDataUpdater.shipData.Speed != "rev80" {
		t.Errorf("Expected speed to be rev80, got %s", mockShipDataUpdater.shipData.Speed)
	}

	adapter.SetSteering("right70")
	time.Sleep(100 * time.Millisecond)

	if mockShipControl.numCmds != 2 {
		t.Fatalf("Expected to receive 2 commands, got %d", mockShipControl.numCmds)
	}
	if mockShipDataUpdater.shipData.Steering != "right70" {
		t.Errorf("Expected steering to be right70, got %s", mockShipDataUpdater.shipData.Steering)
	}
}
