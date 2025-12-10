# WASI Cron POC

Minimal proof-of-concept to run WASM (WASI) jobs on a cron-like schedule using Go.

## Components

- **Scheduler** (`cmd/scheduler`): Go binary that:
  - Loads job definitions from `config/jobs.yaml`
  - Uses `robfig/cron` for schedules (e.g. `@every 10s`, `0 * * * *`)
  - Runs WASI .wasm modules using [wazero](https://github.com/tetratelabs/wazero)

- **WASM job** (`jobs/hello-job`):
  - TinyGo-based job compiled to `wasm32-wasi`
  - Prints a simple message and timestamp

## Prerequisites

- Go 1.22+
- TinyGo (for building the job):
  - Install: `go install github.com/tinygo-org/tinygo@latest`
  - Ensure `tinygo` is on your `PATH`

## Build and run

### 1. Build the WASM job

```bash
cd jobs/hello-job
./build.sh
