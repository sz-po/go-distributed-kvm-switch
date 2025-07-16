package device

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"
)

/* ------------------------------------------------------------------
   Test 1: stdin jest zdublowany – oba potoki dostają ten sam strumień
------------------------------------------------------------------ */

func TestStdioMux_ReadDuplication(t *testing.T) {
	const input = "hello world\nsecond line\n"

	stdin := strings.NewReader(input)
	var stdout bytes.Buffer

	mux := NewStdioMux(stdin, &stdout)
	defer mux.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	var srvData, cliData []byte
	var srvErr, cliErr error

	// czytamy z obu ReadWriterów równolegle
	go func() {
		defer wg.Done()
		srvData, srvErr = io.ReadAll(mux.ServerPipe())
	}()
	go func() {
		defer wg.Done()
		cliData, cliErr = io.ReadAll(mux.ClientPipe())
	}()

	wg.Wait()

	if srvErr != nil {
		t.Fatalf("server pipe read error: %v", srvErr)
	}
	if cliErr != nil {
		t.Fatalf("client pipe read error: %v", cliErr)
	}

	if string(srvData) != input {
		t.Errorf("server received %q, want %q", string(srvData), input)
	}
	if string(cliData) != input {
		t.Errorf("client received %q, want %q", string(cliData), input)
	}
}

/* ------------------------------------------------------------------
   Test 2: równoległe zapisy są serializowane – nie ma przeplatania bajtów
------------------------------------------------------------------ */

func TestStdioMux_ConcurrentWrites(t *testing.T) {
	stdin := strings.NewReader("") // wejście tu nie jest potrzebne
	var stdout bytes.Buffer

	mux := NewStdioMux(stdin, &stdout)
	defer mux.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, _ = mux.ServerPipe().WriteString("server")
		_ = mux.ServerPipe().Flush()
	}()

	go func() {
		defer wg.Done()
		_, _ = mux.ClientPipe().WriteString("client")
		_ = mux.ClientPipe().Flush()
	}()

	wg.Wait()

	got := stdout.String()

	// Dozwolone są dwie poprawne kolejności: "serverclient" lub "clientserver".
	if got != "serverclient" && got != "clientserver" {
		t.Fatalf("unexpected combined output: %q", got)
	}
}

/* ------------------------------------------------------------------
   Test 3: Close() poprawnie zamyka wewnętrzne zasoby
------------------------------------------------------------------ */

func TestStdioMux_Close(t *testing.T) {
	stdin := strings.NewReader("")
	var stdout bytes.Buffer

	mux := NewStdioMux(stdin, &stdout)

	if err := mux.Close(); err != nil {
		t.Fatalf("first Close() returned error: %v", err)
	}

	// drugie Close() może, ale nie musi zwrócić błąd; ważne, by nie panicowało
	_ = mux.Close()
}
