package job

type StateUpdate interface {
	Cur() State
	New() State
}

type State string

const (
	CanceledState  = "canceled"
	CreatedState   = "created"
	ErroredState   = "errored"
	FailedState    = "failed"
	FinishedState  = "finished"
	PassedState    = "passed"
	QueuedState    = "queued"
	ReceivedState  = "received"
	RestartedState = "restarted"
	StartedState   = "started"
)

type serializableStateUpdate struct {
	ID           string                 `json:"id"`
	CurrentState State                  `json:"cur"`
	NewState     State                  `json:"new"`
	State        State                  `json:"state"`
	Meta         map[string]interface{} `json:"meta"`
}

func (ssu *serializableStateUpdate) Cur() State {
	return ssu.CurrentState
}

func (ssu *serializableStateUpdate) New() State {
	return ssu.NewState
}

func NewStateUpdate(jobID string, curState, newState State) StateUpdate {
	return &serializableStateUpdate{
		ID:           jobID,
		CurrentState: curState,
		NewState:     newState,
		State:        newState,
		Meta:         map[string]interface{}{},
	}
}
