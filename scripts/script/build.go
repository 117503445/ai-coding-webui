package main

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

func build() {
	buildFrontend()
	buildBackend()
}

func buildFrontend() {
	feDir := filepath.Join(dirProjectRoot, "fe")

	log.Info().Msg("installing frontend dependencies...")
	if err := runCommand(feDir, "pnpm", "install"); err != nil {
		log.Fatal().Err(err).Msg("pnpm install failed")
	}

	log.Info().Msg("building frontend...")
	if err := runCommand(feDir, "pnpm", "build"); err != nil {
		log.Fatal().Err(err).Msg("frontend build failed")
	}

	distSrc := filepath.Join(feDir, "dist")
	distDst := filepath.Join(dirProjectRoot, "cmd", "rpc", "frontend_dist")

	log.Info().Str("src", distSrc).Str("dst", distDst).Msg("copying frontend dist")
	if err := os.RemoveAll(distDst); err != nil {
		log.Fatal().Err(err).Msg("failed to remove old frontend_dist")
	}
	if err := copyDir(distSrc, distDst); err != nil {
		log.Fatal().Err(err).Msg("failed to copy frontend dist")
	}

	log.Info().Msg("frontend build completed")
}

func buildBackend() {
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

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}
