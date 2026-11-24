package cove

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"
)

const defaultOutput = 1 << 20

var (
	ErrInvalid = errors.New("command is not valid")
	ErrLimited = errors.New("memory limit was not applied")
)

// Request is one command. Memory is a byte cap for the process. Zero means no cap.
type Request struct {
	Path      string
	Args      []string
	Dir       string
	Env       []string
	Timeout   time.Duration
	MaxOutput int
	Memory    uint64
}

// Result is what the process produced. TimedOut means the time limit stopped it.
type Result struct {
	Exit      int    `json:"exit"`
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	TimedOut  bool   `json:"timed_out"`
	Truncated bool   `json:"truncated"`
}

// Run starts Path and waits for it to finish, or for the time limit.
// The child shares a short environment: the Windows system paths plus Env.
func Run(ctx context.Context, req Request) (Result, error) {
	if strings.TrimSpace(req.Path) == "" {
		return Result{}, ErrInvalid
	}
	if req.MaxOutput <= 0 {
		req.MaxOutput = defaultOutput
	}
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, req.Path, req.Args...)
	cmd.Dir = req.Dir
	cmd.Env = baseEnv(req.Env)
	cmd.WaitDelay = 2 * time.Second
	stdout := &limited{max: req.MaxOutput}
	stderr := &limited{max: req.MaxOutput}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	limit, err := prepare(req.Memory)
	if err != nil {
		return Result{}, err
	}
	defer limit.close()
	cmd.Cancel = func() error {
		_ = limit.cancel()
		if cmd.Process != nil {
			return cmd.Process.Kill()
		}
		return nil
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return Result{}, err
	}

	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		return Result{}, err
	}
	if err := limit.assign(cmd.Process.Pid); err != nil {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
		_ = stdin.Close()
		if req.Memory > 0 {
			return Result{}, ErrLimited
		}
	}
	if ctx.Err() != nil {
		_ = limit.cancel()
	}
	_ = stdin.Close()

	var result Result
	err = cmd.Wait()
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()
	result.Truncated = stdout.over || stderr.over
	if ctx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		result.Exit = -1
		return result, nil
	}
	if err == nil {
		return result, nil
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		result.Exit = exit.ExitCode()
		return result, nil
	}
	return result, err
}

func baseEnv(extra []string) []string {
	keep := map[string]struct{}{
		"SYSTEMROOT": {},
		"WINDIR":     {},
		"PATH":       {},
		"COMSPEC":    {},
		"PATHEXT":    {},
	}
	var out []string
	for _, item := range os.Environ() {
		key, _, ok := strings.Cut(item, "=")
		if !ok {
			continue
		}
		if _, want := keep[strings.ToUpper(key)]; want {
			out = append(out, item)
		}
	}
	return append(out, extra...)
}

type limited struct {
	buf  bytes.Buffer
	max  int
	over bool
}

func (w *limited) Write(p []byte) (int, error) {
	room := w.max - w.buf.Len()
	if room <= 0 {
		w.over = true
		return len(p), nil
	}
	if len(p) > room {
		_, _ = w.buf.Write(p[:room])
		w.over = true
		return len(p), nil
	}
	_, _ = w.buf.Write(p)
	return len(p), nil
}

func (w *limited) String() string { return w.buf.String() }
