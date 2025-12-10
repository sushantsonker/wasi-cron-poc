package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/robfig/cron/v3"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"gopkg.in/yaml.v3"
)

type JobConfig struct {
	Name     string            `yaml:"name"`
	WasmPath string            `yaml:"wasm_path"`
	Schedule string            `yaml:"schedule"`
	Env      map[string]string `yaml:"env"`
}

type Config struct {
	Jobs []JobConfig `yaml:"jobs"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal yaml: %w", err)
	}
	return &cfg, nil
}

func main() {
	ctx := context.Background()

	configPath := "config/jobs.yaml"
	if v := os.Getenv("CRON_CONFIG"); v != "" {
		configPath = v
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load config %s: %v", configPath, err)
	}

	if len(cfg.Jobs) == 0 {
		log.Fatalf("no jobs configured in %s", configPath)
	}

	// Create a single wazero runtime we reuse for all jobs.
	runtime := wazero.NewRuntime(ctx)
	defer func() {
		if err := runtime.Close(ctx); err != nil {
			log.Printf("error closing runtime: %v", err)
		}
	}()

	// Instantiate WASI so our modules get stdout/stderr, env, etc.
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, runtime); err != nil {
		log.Fatalf("failed to instantiate WASI: %v", err)
	}

	c := cron.New()

	for _, job := range cfg.Jobs {
		j := job // capture

		log.Printf("registering job %q: wasm=%s schedule=%q", j.Name, j.WasmPath, j.Schedule)

		_, err := c.AddFunc(j.Schedule, func() {
			if err := runWasmJob(ctx, runtime, j); err != nil {
				log.Printf("[job=%s] error: %v", j.Name, err)
			}
		})
		if err != nil {
			log.Fatalf("failed to add job %q to cron: %v", j.Name, err)
		}
	}

	c.Start()
	log.Printf("scheduler started with %d job(s). Press Ctrl+C to exit.", len(cfg.Jobs))

	// Block forever
	select {}
}

func runWasmJob(ctx context.Context, runtime wazero.Runtime, job JobConfig) error {
	wasmBytes, err := os.ReadFile(job.WasmPath)
	if err != nil {
		return fmt.Errorf("read wasm %s: %w", job.WasmPath, err)
	}

	// Compile module (could be cached in a map by path for more efficiency)
	compiled, err := runtime.CompileModule(ctx, wasmBytes)
	if err != nil {
		return fmt.Errorf("compile module: %w", err)
	}
	defer compiled.Close(ctx)

	// Base module config with stdout/stderr
	config := wazero.NewModuleConfig().
		WithStdout(os.Stdout).
		WithStderr(os.Stderr)

	// Add env vars from job.Env
	for k, v := range job.Env {
		config = config.WithEnv(k, v)
	}

	// Instantiate and run `_start` (default for WASI)
	mod, err := runtime.InstantiateModule(ctx, compiled, config)
	if err != nil {
		return fmt.Errorf("instantiate module: %w", err)
	}
	defer mod.Close(ctx)

	return nil
}

