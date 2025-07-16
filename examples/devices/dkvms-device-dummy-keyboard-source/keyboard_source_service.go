package main

import (
	"github.com/sz-po/go-distributed-kvm-switch/pkg/device/api/keyboard/v1_alpha"
)

type KeyboardSourceService struct {
}

func NewKeyboardSourceService() *KeyboardSourceService {
	return &KeyboardSourceService{}
}

func (service *KeyboardSourceService) GetNumLockState() v1_alpha.NumLockState {
	//TODO implement me
	panic("implement me")
}

func (service *KeyboardSourceService) GetCapslockState() v1_alpha.CapslockState {
	//TODO implement me
	panic("implement me")
}

func (service *KeyboardSourceService) GetScrollLockState() v1_alpha.ScrollLockState {
	//TODO implement me
	panic("implement me")
}

func (service *KeyboardSourceService) SubscribeKeyboardEvents() {
	//TODO implement me
	panic("implement me")
}

func (service *KeyboardSourceService) UnsubscribeKeyboardEvents() {
	//TODO implement me
	panic("implement me")
}
