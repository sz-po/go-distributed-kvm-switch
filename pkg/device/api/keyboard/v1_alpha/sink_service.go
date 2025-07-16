package v1_alpha

type SinkService interface {
	SetNumLockState(state NumLockState) error
	SetCapslockState(state CapslockState) error
	SetScrollLockState(state ScrollLockState) error
}
