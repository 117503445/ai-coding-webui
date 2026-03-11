package main

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

func build() {
	outDir := filepath.Join(dirProjectRoot, "data", "rpc")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		log.Fatal().Err(err).Str("dir", outDir).Msg("failed to create output directory")
	}

	info := collectBuildInfo()
	output := filepath.Join(outDir, "rpc")
	ldflags := buildLDFlags(info)

	if err := runCommand(
		dirProjectRoot,
		"go", "build",
		"-trimpath",
		"-o", output,
		"-ldflags", ldflags,
		"./cmd/rpc",
	); err != nil {
		log.Fatal().Err(err).Msg("build failed")
	}

	log.Info().Str("output", output).Msg("build completed")
}
