package ship

import (
	"encoding/json"
	"errors"
	"net"
	"time"

	"github.com/moosethebrown/ship-nav/core"
	"github.com/moosethebrown/ship-nav/core/model"
	"github.com/rs/zerolog"
)

type Configurer interface {
	ShipSocketName() string
	ShipPollingInterval() int64
}

type Adapter struct {
	logger          *zerolog.Logger
	socketName      string
	pollingInterval int64
	shipDataUpdater core.ShipDataUpdater
	stopCh          chan bool
	speedCh         chan string
	steeringCh      chan string
}

func NewAdapter(logger *zerolog.Logger, configurer Configurer,
	shipDataUpdater core.ShipDataUpdater) *Adapter {
	return &Adapter{
		logger:          logger,
		socketName:      configurer.ShipSocketName(),
		pollingInterval: configurer.ShipPollingInterval(),
		shipDataUpdater: shipDataUpdater,
		stopCh:          make(chan bool, 1),
		speedCh:         make(chan string, 1),
		steeringCh:      make(chan string, 1),
	}
}

func (a *Adapter) SetShipDataUpdater(shipDataUpdater core.ShipDataUpdater) {
	a.shipDataUpdater = shipDataUpdater
}

func (a *Adapter) Run() {
	conn, err := net.Dial("unix", a.socketName)
	if err != nil {
		a.logger.Error().Err(err).Msg("Failed to connect to socket")
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(time.Duration(a.pollingInterval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			queryResponse, err := a.query(conn)
			if err != nil {
				a.logger.Error().Err(err).Msg("Failed to send IPCQuery")
				continue
			}
			shipData := &model.ShipData{
				Speed:    queryResponse.Speed,
				Steering: queryResponse.Steering,
			}
			a.shipDataUpdater.UpdateShipData(shipData)
		case speed := <-a.speedCh:
			resp, err := a.command(conn, "set_speed", speed)
			if err != nil {
				a.logger.Error().Err(err).Msg("Failed to send set_speed command")
			}
			if resp.Status != "ok" {
				a.logger.Error().Err(errors.New(resp.Error)).Msg("set_speed command returned error")
			}
		case steering := <-a.steeringCh:
			resp, err := a.command(conn, "set_steering", steering)
			if err != nil {
				a.logger.Error().Err(err).Msg("Failed to send set_steering command")
			}
			if resp.Status != "ok" {
				a.logger.Error().Err(errors.New(resp.Error)).Msg("set_steering command returned error")
			}
		case <-a.stopCh:
			return
		}
	}
}

func (a *Adapter) Stop() {
	a.stopCh <- true
}

func (a *Adapter) SetSpeed(speed string) {
	a.speedCh <- speed
}

func (a *Adapter) SetSteering(steering string) {
	a.steeringCh <- steering
}

func (a *Adapter) query(conn net.Conn) (*IPCQueryResponse, error) {
	req := &IPCRequest{
		Type: "query",
	}
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	_, err = conn.Write(data)
	if err != nil {
		return nil, err
	}

	resp := &IPCQueryResponse{}
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(buf[:n], resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (a *Adapter) command(conn net.Conn, cmd string, data string) (*IPCCommandResponse, error) {
	req := &IPCRequest{
		Type: "cmd",
		Cmd:  cmd,
		Data: data,
	}
	jsonRq, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	_, err = conn.Write(jsonRq)
	if err != nil {
		return nil, err
	}

	resp := &IPCCommandResponse{}
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(buf[:n], resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
