# Cove

Cove runs a command in its own process. The run stops at a time limit, keeps only a capped amount of output, and can refuse a process that passes a memory cap. On Windows the process sits in a job, so a child it starts is closed with it.

The demo starts itself once and prints:

```text
cove
```

## Run

```powershell
go test ./...
go test -race -count=1 ./...
go run ./cmd/demo
```

```go
result, err := cove.Run(ctx, cove.Request{
    Path:    "tool.exe",
    Args:    []string{"--check"},
    Timeout: 5 * time.Second,
    Memory:  32 << 20,
})
```

`TimedOut` means the time limit won. `Truncated` means the output was longer than `MaxOutput` (the default is 1 MiB). `Memory` is the job's byte cap. The child receives the Windows system path variables and any variables in `Env`.
