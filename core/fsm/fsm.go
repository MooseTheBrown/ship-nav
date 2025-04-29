package fsm

type StateHandler[E any] interface {
	OnEnter()
	OnExit()
	HandleEvent(event E) string
}

type State[E any] struct {
	handler     StateHandler[E]
	transitions map[string]string
}

func NewState[E any](handler StateHandler[E], transitions map[string]string) *State[E] {
	return &State[E]{
		handler:     handler,
		transitions: transitions,
	}
}

type FSM[E any] struct {
	states       map[string]*State[E]
	currentState string
}

func NewFSM[E any](states map[string]*State[E], initialState string) *FSM[E] {
	fsm := &FSM[E]{
		states:       states,
		currentState: initialState,
	}
	fsm.states[fsm.currentState].handler.OnEnter()
	return fsm
}

func (fsm *FSM[E]) HandleEvent(event E) {
	currentState := fsm.states[fsm.currentState]
	transition := currentState.handler.HandleEvent(event)
	if transition != "" {
		targetState, ok := currentState.transitions[transition]
		if ok {
			currentState.handler.OnExit()
			fsm.currentState = targetState
			currentState = fsm.states[fsm.currentState]
			currentState.handler.OnEnter()
		}
	}
}

func (fsm *FSM[E]) CurrentState() string {
	return fsm.currentState
}
