package cove

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	switch os.Getenv("COVE_HELPER") {
	case "echo":
		fmt.Print("cove")
		os.Exit(0)
	case "code":
		os.Exit(7)
	case "tick":
		path := os.Args[len(os.Args)-1]
		for {
			f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if err == nil {
				_, _ = f.Write([]byte("x"))
				_ = f.Close()
			}
			time.Sleep(40 * time.Millisecond)
		}
	case "loud":
		fmt.Print("abcdefghijklmnopqrstuvwxyz")
		os.Exit(0)
	case "alloc":
		buf := make([]byte, 1)
		_, _ = os.Stdin.Read(buf)
		block := make([]byte, 128<<20)
		for i := range block {
			block[i] = 1
		}
		fmt.Print(len(block))
		os.Exit(0)
	default:
		os.Exit(m.Run())
	}
}

func helper(env string, args ...string) Request {
	return Request{
		Path:    os.Args[0],
		Args:    args,
		Env:     []string{"COVE_HELPER=" + env},
		Timeout: 5 * time.Second,
	}
}

func TestEcho(t *testing.T) {
	got, err := Run(context.Background(), helper("echo"))
	if err != nil {
		t.Fatal(err)
	}
	if got.Exit != 0 || got.Stdout != "cove" || got.TimedOut {
		t.Fatalf("%+v", got)
	}
}

func TestExitCode(t *testing.T) {
	got, err := Run(context.Background(), helper("code"))
	if err != nil {
		t.Fatal(err)
	}
	if got.Exit != 7 {
		t.Fatalf("exit %d", got.Exit)
	}
}

func TestTimeoutStopsTheProcess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ticks")
	req := helper("tick", path)
	req.Timeout = 200 * time.Millisecond
	got, err := Run(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !got.TimedOut {
		t.Fatalf("%+v", got)
	}
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond)
	second, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(second) != len(first) || len(first) == 0 {
		t.Fatalf("ticks %d then %d", len(first), len(second))
	}
}

func TestOutputCap(t *testing.T) {
	req := helper("loud")
	req.MaxOutput = 4
	got, err := Run(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if got.Stdout != "abcd" || !got.Truncated {
		t.Fatalf("%+v", got)
	}
}

func TestRejectsEmptyCommand(t *testing.T) {
	if _, err := Run(context.Background(), Request{}); err != ErrInvalid {
		t.Fatalf("%v", err)
	}
}
