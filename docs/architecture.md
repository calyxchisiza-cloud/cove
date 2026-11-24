# Architecture

Cove starts one process and waits for it. Three limits apply.

The time limit comes from the request, or from the caller's context when that ends first. When it fires, Cove terminates the Windows job and the process. `TimedOut` is set and the exit code is reported as -1. Output written before the stop is kept.

The output limit keeps the first `MaxOutput` bytes of standard output and the first `MaxOutput` bytes of standard error. Further writes are discarded and `Truncated` is set.

The memory limit is a Windows job object. The job's process-memory and job-memory caps are the same value. `JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE` is set, so closing the job closes every process assigned to it. The process is assigned to the job after `Start` returns and before its standard input is closed. A child that waits to read standard input is still under the cap when it continues. If the job cannot be created, a request that asked for a memory cap fails instead of running without one.

The environment passed to the child is `SystemRoot`, `WinDir`, `Path`, `ComSpec`, `Pathext`, and the variables on the request.
