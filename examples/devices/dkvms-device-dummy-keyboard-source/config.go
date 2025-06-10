package main

import "time"

type RandomKeystrokeConfig struct {
	Enabled  bool `default:"false"`
	Interval struct {
		From time.Duration
		To   time.Duration
	}
	Characters struct {
		SmallLetters bool
		BigLetters   bool
		Numbers      bool
		Symbols      bool
		WhiteSpace   bool
	}
}

type Config struct {
	RandomKeystroke RandomKeystrokeConfig `json:"randomKeystroke"`
}
