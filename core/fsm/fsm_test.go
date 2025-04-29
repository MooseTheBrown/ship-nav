package fsm

import "testing"

// mockStateHandler is a mock implementation of the StateHandler interface for testing.
type mockStateHandler[E any] struct {
	OnEnterCalls     int
	OnExitCalls      int
	HandleEventCall  E
	ReturnTransition string
}

func (m *mockStateHandler[E]) OnEnter() {
	m.OnEnterCalls++
}

func (m *mockStateHandler[E]) OnExit() {
	m.OnExitCalls++
}

func (m *mockStateHandler[E]) HandleEvent(event E) string {
	m.HandleEventCall = event
	return m.ReturnTransition
}

// TestNewState tests the NewState function.
func TestNewState(t *testing.T) {
	handler := &mockStateHandler[string]{}
	transitions := map[string]string{"initDone": "nextState"}
	state := NewState(handler, transitions)

	if state == nil {
		t.Error("Expected state to be created but got nil")
	}
	if len(state.transitions) != 1 || (state.transitions["initDone"] != "nextState") {
		t.Error("Expected transitions map to contain 'nextState' transition")
	}
}

// TestNewFSM tests the NewFSM function.
func TestNewFSM(t *testing.T) {
	initialStateHandler := &mockStateHandler[string]{}
	states := map[string]*State[string]{
		"initial": {
			handler:     initialStateHandler,
			transitions: map[string]string{"initDone": "nextState"},
		},
	}
	fsm := NewFSM(states, "initial")

	if fsm.CurrentState() != "initial" {
		t.Error("Expected fsm.CurrentState() to return 'initia' but got", fsm.CurrentState())
	}

	if initialStateHandler.OnEnterCalls != 1 {
		t.Errorf("Expected 1 OnEnter call in initial state, got %d", initialStateHandler.OnEnterCalls)
	}
	if initialStateHandler.OnExitCalls != 0 {
		t.Errorf("Expected 0 OnExit calls in initial state, got %d", initialStateHandler.OnExitCalls)
	}
}

// TestHandleEvent tests the HandleEvent function.
func TestHandleEvent(t *testing.T) {
	initialStateHandler := &mockStateHandler[string]{
		ReturnTransition: "initDone",
	}
	nextStateHandler := &mockStateHandler[string]{
		ReturnTransition: "backToInitial",
	}
	states := map[string]*State[string]{
		"initial": {
			handler:     initialStateHandler,
			transitions: map[string]string{"initDone": "nextState"},
		},
		"nextState": {
			handler:     nextStateHandler,
			transitions: map[string]string{"backToInitial": "initial"},
		},
	}
	fsm := NewFSM(states, "initial")

	// Test handling an event that triggers a transition
	event := "someEvent"
	fsm.HandleEvent(event)
	if fsm.currentState != "nextState" {
		t.Error("Expected state to be 'nextState' after handling event but got", fsm.currentState)
	}
	if initialStateHandler.OnEnterCalls != 1 {
		t.Errorf("Expected 1 OnEnter call in initial state, got %d", initialStateHandler.OnEnterCalls)
	}
	if initialStateHandler.OnExitCalls != 1 {
		t.Errorf("Expected 1 OnExit call in initial state, got %d", initialStateHandler.OnExitCalls)
	}
	if nextStateHandler.OnEnterCalls != 1 {
		t.Errorf("Expected 1 OnEnter call in next state, got %d", nextStateHandler.OnEnterCalls)
	}
	if nextStateHandler.OnExitCalls != 0 {
		t.Errorf("Expected 0 OnExit calls in next state, got %d", nextStateHandler.OnExitCalls)
	}

	// Test returning to initial state
	event = "otherEvent"
	fsm.HandleEvent(event)
	if fsm.currentState != "initial" {
		t.Error("Expected state to be 'initial' after handling event but got", fsm.currentState)
	}
	if initialStateHandler.OnEnterCalls != 2 {
		t.Errorf("Expected 2 OnEnter calls in initial state, got %d", initialStateHandler.OnEnterCalls)
	}
	if initialStateHandler.OnExitCalls != 1 {
		t.Errorf("Expected 1 OnExit call in initial state, got %d", initialStateHandler.OnExitCalls)
	}
	if nextStateHandler.OnEnterCalls != 1 {
		t.Errorf("Expected 1 OnEnter call in next state, got %d", nextStateHandler.OnEnterCalls)
	}
	if nextStateHandler.OnExitCalls != 1 {
		t.Errorf("Expected 0 OnExit calls in next state, got %d", nextStateHandler.OnExitCalls)
	}
}
