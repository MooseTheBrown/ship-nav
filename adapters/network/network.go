package network

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"github.com/google/uuid"
	"github.com/moosethebrown/ship-nav/core"
	"github.com/moosethebrown/ship-nav/core/model"
	"github.com/rs/zerolog"
)

type Adapter struct {
	logger                *zerolog.Logger
	socketName            string
	listener              net.Listener
	conns                 map[string]net.Conn
	shipDataProvider      core.ShipDataProvider
	positionDataProvider  core.PositionDataProvider
	waypointsDataProvider core.WaypointDataProvider
	navController         core.NavigationController
	waypointsUpdater      core.WaypointsUpdater
}

func NewAdapter(socketName string, sp core.ShipDataProvider,
	pp core.PositionDataProvider, wp core.WaypointDataProvider,
	nc core.NavigationController, wu core.WaypointsUpdater, logger *zerolog.Logger) *Adapter {
	return &Adapter{
		socketName:            socketName,
		conns:                 make(map[string]net.Conn),
		shipDataProvider:      sp,
		positionDataProvider:  pp,
		waypointsDataProvider: wp,
		navController:         nc,
		waypointsUpdater:      wu,
		logger:                logger,
	}
}

func (a *Adapter) Run() {
	defer a.handlePanic()

	var err error
	a.listener, err = net.Listen("unix", a.socketName)
	if err != nil {
		a.logger.Error().Err(err).Msgf("Failed to open socket %s for listening", a.socketName)
		return
	}

	for {
		conn, err := a.listener.Accept()
		if err != nil {
			a.logger.Error().Err(err).Msgf("Failed to accept connection")
			break
		}
		clientId := uuid.NewString()
		a.conns[clientId] = conn
		go a.handleClient(clientId)
	}
}

func (a *Adapter) Stop() {
	a.logger.Info().Msg("Stopping")

	if a.listener != nil {
		a.listener.Close()
	}

	for clientId, conn := range a.conns {
		a.logger.Debug().Msgf("Closing connection for client %s", clientId)
		conn.Close()
	}
}

func (a *Adapter) handlePanic() {
	if a.listener != nil {
		a.listener.Close()
	}

	what := recover()
	if what == nil {
		return
	}

	switch what.(type) {
	case string:
		err := errors.New(what.(string))
		a.logger.Error().Err(err).Msg("Recovered from panic, restarting")
	case fmt.Stringer:
		err := errors.New(what.(fmt.Stringer).String())
		a.logger.Error().Err(err).Msg("Recovered from panic, restarting")
	default:
		a.logger.Error().Msg("Recovered from panic, restarting")
	}

	go a.Run()
}

func (a *Adapter) handleClient(clientId string) {
	a.logger.Info().Msgf("Connected client %s", clientId)

	conn, ok := a.conns[clientId]
	if !ok {
		a.logger.Error().Msgf("Attempted to handle non-existent client with ID '%s'", clientId)
		return
	}
	defer conn.Close()

	for {
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err != nil {
			a.logger.Error().Err(err).Msgf("Error reading from client %s",
				clientId)
			break
		}

		var rq Request
		err = json.Unmarshal(buf[:n], &rq)
		if err != nil {
			a.logger.Error().Err(err).Msg("Error unmarshalling request")
			break
		}

		resp, err := a.handleRequest(&rq)
		if err != nil {
			a.logger.Error().Err(err).Msg("Failed to process request")
			break
		}

		_, err = conn.Write(resp)
		if err != nil {
			a.logger.Error().Err(err).Msg("Failed to send response")
			break
		}
	}
}

func (a *Adapter) handleRequest(rq *Request) ([]byte, error) {
	if rq.Type == rqTypeQuery {
		return a.handleQuery()
	} else if rq.Type == rqTypeCmd {
		return a.handleCommand(rq)
	} else {
		return nil, errors.New(fmt.Sprintf("Invalid request type %s", rq.Type))
	}
}

func (a *Adapter) handleQuery() ([]byte, error) {
	var resp QueryResponse
	bearing, position := a.positionDataProvider.GetPositionData()
	resp.PositionData = &PositionData{
		Angle:         bearing.AngleDeg(),
		NumSatellites: position.NumSatellites,
		Latitude:      position.Latitude,
		Longitude:     position.Longitude,
		SpeedKnots:    position.SpeedKnots,
		SpeedKm:       position.SpeedKm,
	}

	shipData := a.shipDataProvider.GetShipData()
	resp.ShipData = &ShipData{
		Speed:    shipData.Speed,
		Steering: shipData.Steering,
	}

	waypoints := a.waypointsDataProvider.GetWaypoints()
	resp.Waypoints = make([]*Waypoint, len(waypoints))
	for i, waypoint := range waypoints {
		resp.Waypoints[i] = &Waypoint{
			Latitude:  waypoint.Latitude,
			Longitude: waypoint.Longitude,
		}
	}

	respData, err := json.Marshal(resp)
	return respData, err
}

func (a *Adapter) handleCommand(rq *Request) ([]byte, error) {
	resp := &CommandResponse{
		Status: "ok",
	}

	switch rq.Cmd {
	case cmdNavStart:
		a.navController.StartNavigation()
	case cmdNavStop:
		a.navController.StopNavigation()
	case cmdNetLoss:
		a.navController.NetworkLost()
	case cmdSetWaypoints:
		if rq.Waypoints == nil || len(rq.Waypoints) == 0 {
			resp.Status = "failure"
			resp.Error = "no waypoints provided"
			break
		}
		wps := make([]*model.Waypoint, len(rq.Waypoints))
		for i, wp := range rq.Waypoints {
			wps[i] = &model.Waypoint{
				Latitude:  wp.Latitude,
				Longitude: wp.Longitude,
			}
		}
		a.waypointsUpdater.SetWaypoints(wps)
	case cmdAddWaypoint:
		if rq.Waypoints == nil || len(rq.Waypoints) == 0 {
			resp.Status = "failure"
			resp.Error = "waypoint is not provided"
			break
		}
		wp := &model.Waypoint{
			Latitude:  rq.Waypoints[0].Latitude,
			Longitude: rq.Waypoints[0].Longitude,
		}
		a.waypointsUpdater.AddWaypoint(wp)
	case cmdClearWaypoints:
		a.waypointsUpdater.ClearWaypoints()
	case cmdSetHomeWaypoint:
		if rq.Waypoints == nil || len(rq.Waypoints) == 0 {
			resp.Status = "failure"
			resp.Error = "waypoint is not provided"
			break
		}
		wp := &model.Waypoint{
			Latitude:  rq.Waypoints[0].Latitude,
			Longitude: rq.Waypoints[0].Longitude,
		}
		a.waypointsUpdater.SetHomeWaypoint(wp)
	}

	respData, err := json.Marshal(resp)
	return respData, err
}
