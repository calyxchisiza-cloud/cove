//go:build windows

package cove

import (
	"context"
	"testing"
	"time"
)

func TestMemoryCapStopsTheProcess(t *testing.T) {
	req := helper("alloc")
	req.Memory = 32 << 20
	req.Timeout = 15 * time.Second
	got, err := Run(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if got.Stdout != "" || got.Exit == 0 {
		t.Fatalf("allocation was allowed: %+v", got)
	}
}
