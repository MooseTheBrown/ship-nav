package position

import (
	"encoding/json"
	"errors"
	"net"
	"time"

	"github.com/rs/zerolog"

	"github.com/moosethebrown/ship-nav/core"
	"github.com/moosethebrown/ship-nav/core/model"
)

type Configurer interface {
	PositionSocketName() string
	PositionPollingInterval() int64
	Declination() float64
}

type Adapter struct {
	logger          *zerolog.Logger
	socketName      string
	pollingInterval int64
	stopCh          chan bool
	calibrating     bool
	calibrationCh   chan bool
	positionUpdater core.PositionUpdater
	bearingUpdater  core.BearingUpdater
	declination     float64
}

func NewAdapter(logger *zerolog.Logger, configurer Configurer,
	positionUpdater core.PositionUpdater, bearingUpdater core.BearingUpdater) *Adapter {
	return &Adapter{
		logger:          logger,
		socketName:      configurer.PositionSocketName(),
		pollingInterval: configurer.PositionPollingInterval(),
		stopCh:          make(chan bool, 1),
		calibrationCh:   make(chan bool, 1),
		positionUpdater: positionUpdater,
		bearingUpdater:  bearingUpdater,
		declination:     configurer.Declination(),
	}
}

func (a *Adapter) Run() {
	conn, err := net.Dial("unix", a.socketName)
	if err != nil {
		a.logger.Error().Err(err).Msg("Failed to open Unix socket")
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(time.Duration(a.pollingInterval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if a.calibrating {
				continue
			}

			gpsInfo, err := a.gpsInfoRequest(conn)
			if err != nil {
				a.logger.Error().Err(err).Msg("Failed to query gps info")
				continue
			}

			position := &model.Position{
				NumSatellites: int8(gpsInfo.NumSatellites),
				Latitude:      gpsInfo.Latitude,
				Longitude:     gpsInfo.Longitude,
				SpeedKnots:    gpsInfo.SpeedKnots,
				SpeedKm:       gpsInfo.SpeedKm,
			}
			a.positionUpdater.UpdatePosition(position)

			magnetometerInfo, err := a.magnetometerInfoRequest(conn)
			if err != nil {
				a.logger.Error().Err(err).Msg("Failed to query magnetometer info")
			}
			bearing := model.NewBearing(a.declination)
			bearing.SetInt(magnetometerInfo.X, magnetometerInfo.Y)
			a.bearingUpdater.UpdateBearing(bearing)

		case <-a.stopCh:
			return

		case a.calibrating = <-a.calibrationCh:
			if a.calibrating {
				a.logger.Info().Msg("Starting calibration")
			} else {
				a.logger.Info().Msg("Stopping calibration")
			}
			resp, err := a.calibrationRequest(conn, a.calibrating)
			if err != nil {
				a.logger.Error().Err(err).Msg("Failed to start/stop calibration")
			}
			if !resp.Success {
				a.logger.Error().Msg("ship-position failed to start/stop calibration")
			}
		}
	}
}

func (a *Adapter) Stop() {
	a.stopCh <- true
}

func (a *Adapter) StartCalibration() {
	a.calibrationCh <- true
}

func (a *Adapter) StopCalibration() {
	a.calibrationCh <- false
}

func (a *Adapter) gpsInfoRequest(conn net.Conn) (*GPSInfoResponse, error) {
	rq := &IPCRequest{Cmd: CmdGetGPS}
	data, err := json.Marshal(rq)
	if err != nil {
		return nil, err
	}
	_, err = conn.Write(data)
	if err != nil {
		return nil, err
	}

	resp := &GPSInfoResponse{}
	errResp := &ErrorResponse{}
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(buf[:n], &resp)
	if err != nil {
		err = json.Unmarshal(buf[:n], &errResp)
		if err == nil {
			return nil, errors.New(errResp.ErrorMessage)
		} else {
			return nil, err
		}
	}

	return resp, nil
}

func (a *Adapter) magnetometerInfoRequest(conn net.Conn) (*MagnetometerInfoResponse, error) {
	rq := &IPCRequest{Cmd: CmdGetMagnetometer}
	data, err := json.Marshal(rq)
	if err != nil {
		return nil, err
	}
	_, err = conn.Write(data)
	if err != nil {
		return nil, err
	}

	resp := &MagnetometerInfoResponse{}
	errResp := &ErrorResponse{}
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(buf[:n], &resp)
	if err != nil {
		err = json.Unmarshal(buf[:n], &errResp)
		if err == nil {
			return nil, errors.New(errResp.ErrorMessage)
		} else {
			return nil, err
		}
	}
	return resp, nil
}

func (a *Adapter) calibrationRequest(conn net.Conn, start bool) (*CalibrationResponse, error) {
	rq := &IPCRequest{}
	if start {
		rq.Cmd = CmdStartCalibration
	} else {
		rq.Cmd = CmdStopCalibration
	}
	data, err := json.Marshal(rq)
	if err != nil {
		return nil, err
	}
	_, err = conn.Write(data)
	if err != nil {
		return nil, err
	}

	resp := &CalibrationResponse{}
	errResp := &ErrorResponse{}
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(buf[:n], &resp)
	if err != nil {
		err = json.Unmarshal(buf[:n], &errResp)
		if err == nil {
			return nil, errors.New(errResp.ErrorMessage)
		} else {
			return nil, err
		}
	}
	return resp, nil
}
