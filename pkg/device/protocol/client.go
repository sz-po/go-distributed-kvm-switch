package protocol

import "io"

type Client struct {
	pipe io.ReadWriter
}

func NewClient(pipe io.ReadWriter) *Client {
	return &Client{
		pipe: pipe,
	}
}
