package core

import (
	"time"

	"github.com/moosethebrown/ship-nav/core/fsm"
	"github.com/moosethebrown/ship-nav/core/model"
	"github.com/rs/zerolog"
)

const (
	defaultUpdateBufSize = 1024
)

type Configurer interface {
	Declination() float64
	UpdateBufSize() int
	TurningSpeed() string
	TurningSteeringLeft() string
	TurningSteeringRight() string
	ApproachSpeed() string
	FullSpeed() string
	ApproachDistance() float64
	DistanceInaccuracy() float64
}

const (
	waypointCmdSet = iota
	waypointCmdAdd
	waypointCmdClear
)

type waypointsCmd struct {
	cmd uint8
	arg []*model.Waypoint
}

type coreData struct {
	declination   float64
	position      *model.Position
	homeWaypoint  *model.Waypoint
	curBearing    *model.Bearing
	targetBearing *model.Bearing
	shipData      *model.ShipData
	waypoints     *model.Waypoints
}

type Core struct {
	data           *coreData
	positionCh     chan *model.Position
	homeWaypointCh chan *model.Waypoint
	bearingCh      chan *model.Bearing
	shipDataCh     chan *model.ShipData
	waypointsCh    chan *waypointsCmd
	navCh          chan bool
	netLossCh      chan bool
	stopCh         chan bool
	fsm            *fsm.FSM[Event]
	logger         *zerolog.Logger
}

func NewCore(configurer Configurer, shipControl ShipControl, logger *zerolog.Logger) *Core {
	updateBufSize := configurer.UpdateBufSize()
	if updateBufSize == 0 {
		updateBufSize = defaultUpdateBufSize
	}

	coreData := &coreData{
		declination:   configurer.Declination(),
		position:      &model.Position{},
		curBearing:    model.NewBearing(configurer.Declination()),
		targetBearing: model.NewBearing(configurer.Declination()),
		shipData:      &model.ShipData{},
		waypoints:     model.NewWaypoints(),
	}

	idleLogger := logger.With().Str("state", "idle").Logger()
	turningLogger := logger.With().Str("state", "turning").Logger()
	movingLogger := logger.With().Str("state", "moving").Logger()
	turningHomeLogger := logger.With().Str("state", "turning home").Logger()
	movingHomeLogger := logger.With().Str("state", "moving home").Logger()
	stoppingLogger := logger.With().Str("state", "stopping").Logger()

	idleHandler := newIdleHandler(&idleLogger, coreData)
	turningHandler := newTurningHandler(&turningLogger, coreData, shipControl,
		configurer.TurningSpeed(), configurer.TurningSteeringLeft(),
		configurer.TurningSteeringRight())
	movingHandler := newMovingHandler(&movingLogger, coreData, shipControl, configurer.ApproachSpeed(),
		configurer.FullSpeed(), configurer.ApproachDistance(), configurer.DistanceInaccuracy())
	turningHomeHandler := newTurningHomeHandler(&turningHomeLogger, coreData, shipControl,
		configurer.TurningSpeed(), configurer.TurningSteeringLeft(), configurer.TurningSteeringRight())
	movingHomeHandler := newMovingHomeHandler(&movingHomeLogger, coreData, shipControl,
		configurer.ApproachSpeed(), configurer.FullSpeed(), configurer.ApproachDistance(),
		configurer.DistanceInaccuracy())
	stoppingHandler := newStoppingHandler(&stoppingLogger, coreData, shipControl)

	return &Core{
		data:           coreData,
		positionCh:     make(chan *model.Position, updateBufSize),
		homeWaypointCh: make(chan *model.Waypoint, updateBufSize),
		bearingCh:      make(chan *model.Bearing, updateBufSize),
		shipDataCh:     make(chan *model.ShipData, updateBufSize),
		waypointsCh:    make(chan *waypointsCmd, updateBufSize),
		navCh:          make(chan bool, updateBufSize),
		netLossCh:      make(chan bool, updateBufSize),
		stopCh:         make(chan bool, 1),
		fsm: fsm.NewFSM(map[string]*fsm.State[Event]{
			"idle": fsm.NewState(idleHandler, map[string]string{
				"nav start":     "turning",
				"net loss home": "turning home",
			}),
			"turning": fsm.NewState(turningHandler, map[string]string{
				"nav stop":          "idle",
				"bearing adjust":    "moving",
				"net loss stop":     "stopping",
				"waypoints cleared": "stopping",
				"net loss home":     "turning home",
			}),
			"moving": fsm.NewState(movingHandler, map[string]string{
				"nav stop":          "idle",
				"waypoint":          "turning",
				"waypoints set":     "turning",
				"last waypoint":     "stopping",
				"net loss stop":     "stopping",
				"waypoints cleared": "stopping",
				"net loss home":     "turning home",
			}),
			"turning home": fsm.NewState(turningHomeHandler, map[string]string{
				"nav stop":       "idle",
				"bearing adjust": "moving home",
			}),
			"moving home": fsm.NewState(movingHomeHandler, map[string]string{
				"nav stop":     "idle",
				"home reached": "stopping",
			}),
			"stopping": fsm.NewState(stoppingHandler, map[string]string{
				"ship stopped": "idle",
			}),
		}, "idle"),
		logger: logger,
	}
}

func (c *Core) UpdatePosition(position *model.Position) {
	if position != nil {
		c.positionCh <- position
	}
}

func (c *Core) UpdateBearing(bearing *model.Bearing) {
	if bearing != nil {
		c.bearingCh <- bearing
	}
}

func (c *Core) UpdateShipData(shipData *model.ShipData) {
	if shipData != nil {
		c.shipDataCh <- shipData
	}
}

func (c *Core) SetWaypoints(waypoints []*model.Waypoint) {
	if len(waypoints) > 0 {
		c.waypointsCh <- &waypointsCmd{
			cmd: waypointCmdSet,
			arg: waypoints,
		}
	}
}

func (c *Core) AddWaypoint(waypoint *model.Waypoint) {
	if waypoint != nil {
		arg := make([]*model.Waypoint, 1)
		arg[0] = waypoint
		c.waypointsCh <- &waypointsCmd{
			cmd: waypointCmdAdd,
			arg: arg,
		}
	}
}

func (c *Core) ClearWaypoints() {
	c.waypointsCh <- &waypointsCmd{
		cmd: waypointCmdClear,
		arg: nil,
	}
}

func (c *Core) StartNavigation() {
	c.navCh <- true
}

func (c *Core) StopNavigation() {
	c.navCh <- false
}

func (c *Core) SetHomeWaypoint(homeWaypoint *model.Waypoint) {
	c.homeWaypointCh <- homeWaypoint
}

func (c *Core) NetworkLost() {
	c.netLossCh <- true
}

func (c *Core) Run() {
	ticker := time.NewTicker(time.Duration(3 * time.Second))
	defer ticker.Stop()

core_loop:
	for {
		evt := Event(eventUndefined)
		select {
		case <-ticker.C:
			c.logger.Info().Msgf("current state = %s", c.fsm.CurrentState())
		case newPosition := <-c.positionCh:
			c.data.position = newPosition
			evt = eventPositionUpdate
		case newHomeWaypoint := <-c.homeWaypointCh:
			c.data.homeWaypoint = newHomeWaypoint
			evt = eventHomeWaypointUpdate
		case newBearing := <-c.bearingCh:
			c.data.curBearing = newBearing
			evt = eventBearingUpdate
		case newShipData := <-c.shipDataCh:
			c.data.shipData = newShipData
			evt = eventShipDataUpdate
		case waypointCmd := <-c.waypointsCh:
			evt = c.handleWaypointsCmd(waypointCmd)
		case startNav := <-c.navCh:
			if startNav {
				evt = eventNavStart
			} else {
				evt = eventNavStop
			}
		case netLoss := <-c.netLossCh:
			if netLoss {
				evt = eventNetLoss
			}
		case <-c.stopCh:
			break core_loop
		}
		c.fsm.HandleEvent(evt)
	}
}

func (c *Core) Stop() {
	c.stopCh <- true
}

func (c *Core) GetPositionData() (*model.Bearing, *model.Position) {
	var bearing model.Bearing
	bearing = *c.data.curBearing

	var position model.Position
	position = *c.data.position

	return &bearing, &position
}

func (c *Core) GetShipData() *model.ShipData {
	var shipData model.ShipData
	shipData = *c.data.shipData
	return &shipData
}

func (c *Core) GetWaypoints() []*model.Waypoint {
	waypoints := make([]*model.Waypoint, 0)
	for waypoint := c.data.waypoints.GetNextWaypoint(); waypoint != nil; waypoint = c.data.waypoints.GetNextWaypoint() {
		waypoints = append(waypoints, waypoint)
	}
	return waypoints
}

func (c *Core) handleWaypointsCmd(cmd *waypointsCmd) Event {
	switch cmd.cmd {
	case waypointCmdSet:
		c.data.waypoints.SetWaypoints(cmd.arg)
		return eventWaypointsSet
	case waypointCmdAdd:
		if len(cmd.arg) > 0 {
			c.data.waypoints.AddWaypoint(cmd.arg[0])
			return eventWaypointAdded
		}
	case waypointCmdClear:
		c.data.waypoints = model.NewWaypoints()
		return eventWaypointsCleared
	}

	return eventUndefined
}
