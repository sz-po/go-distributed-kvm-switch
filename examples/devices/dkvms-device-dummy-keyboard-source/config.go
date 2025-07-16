package main

import (
	"github.com/sz-po/go-distributed-kvm-switch/pkg/utils"
)

type RandomKeystrokeConfig struct {
	Enabled  bool `default:"false" validate:"required"`
	Interval struct {
		From utils.Duration `default:"1s"`
		To   utils.Duration `default:"5s"`
	} `validate:"required"`
	Characters struct {
		SmallLetters bool `default:"true"`
		BigLetters   bool `default:"true"`
		Numbers      bool `default:"true"`
		Symbols      bool `default:"false"`
		WhiteSpace   bool `default:"false"`
	}
}

type Config struct {
	RandomKeystroke RandomKeystrokeConfig `json:"randomKeystroke" validate:"required"`
}
