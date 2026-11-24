package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"cove"
)

func main() {
	if os.Getenv("COVE_DEMO") == "1" {
		fmt.Print("cove")
		return
	}
	result, err := cove.Run(context.Background(), cove.Request{
		Path:    os.Args[0],
		Env:     []string{"COVE_DEMO=1"},
		Timeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(result.Stdout)
}
