package core

import (
	"github.com/rs/zerolog"
)

type idleHandler struct {
	logger   *zerolog.Logger
	coreData *coreData
}

func newIdleHandler(logger *zerolog.Logger, coreData *coreData) *idleHandler {
	return &idleHandler{
		logger:   logger,
		coreData: coreData,
	}
}

func (handler *idleHandler) OnEnter() {
	handler.logger.Debug().Msg("OnEnter")
}

func (handler *idleHandler) OnExit() {
	handler.logger.Debug().Msg("OnExit")
}

func (handler *idleHandler) HandleEvent(event Event) string {
	handler.logger.Debug().Msgf("HandleEvent event=%s", event.String())
	switch event {
	case eventNavStart:
		handler.logger.Info().Msg("nav start")
		return "nav start"
	case eventNetLoss:
		if handler.coreData.homeWaypoint != nil {
			handler.logger.Info().Msg("net loss home")
			return "net loss home"
		}
	}
	return ""
}
