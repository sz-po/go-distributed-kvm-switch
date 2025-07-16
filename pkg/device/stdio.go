package device

import (
	"bufio"
	"io"
	"sync"
)

type StdioMux struct {
	serverPipe *bufio.ReadWriter
	clientPipe *bufio.ReadWriter

	wSrv *io.PipeWriter
	wCli *io.PipeWriter
}

func NewStdioMux(stdin io.Reader, stdout io.Writer) *StdioMux {
	serverReader, serverWriter := io.Pipe()
	clientReader, clientWriter := io.Pipe()

	go func() {
		defer serverWriter.Close()
		defer clientWriter.Close()
		_, _ = io.Copy(io.MultiWriter(serverWriter, clientWriter), stdin)
	}()

	var mu sync.Mutex
	safeStdout := &lockedWriter{w: stdout, mu: &mu}

	return &StdioMux{
		serverPipe: bufio.NewReadWriter(
			bufio.NewReader(serverReader),
			bufio.NewWriter(safeStdout),
		),
		clientPipe: bufio.NewReadWriter(
			bufio.NewReader(clientReader),
			bufio.NewWriter(safeStdout),
		),
		wSrv: serverWriter,
		wCli: clientWriter,
	}
}

func (m *StdioMux) ServerPipe() *bufio.ReadWriter { return m.serverPipe }

func (m *StdioMux) ClientPipe() *bufio.ReadWriter { return m.clientPipe }

func (m *StdioMux) Close() error {
	err1 := m.wSrv.Close()
	err2 := m.wCli.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

type lockedWriter struct {
	w  io.Writer
	mu *sync.Mutex
}

func (l *lockedWriter) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w.Write(p)
}
