package v1_alpha

type SourceService interface {
	GetNumLockState() NumLockState
	GetCapslockState() CapslockState
	GetScrollLockState() ScrollLockState
	SubscribeKeyboardEvents()
	UnsubscribeKeyboardEvents()
}
