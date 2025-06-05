package main

type DisplayConfig struct {
	Id string `env:"ID"`
}

type Config struct {
	Display DisplayConfig `env:", prefix=DISPLAY_"`
}
