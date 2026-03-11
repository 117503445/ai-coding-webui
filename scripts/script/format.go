package main

import (
	"path/filepath"

	"github.com/rs/zerolog/log"
)

func format() {
	scriptDir := filepath.Join(dirProjectRoot, "scripts", "script")

	steps := []struct {
		dir  string
		name string
		args []string
	}{
		{dirProjectRoot, "go", []string{"mod", "tidy"}},
		{scriptDir, "go", []string{"mod", "tidy"}},
		{dirProjectRoot, "go", []string{"fmt", "./..."}},
		{scriptDir, "go", []string{"fmt", "./..."}},
		{dirProjectRoot, "buf", []string{"format", "-w"}},
	}

	for _, step := range steps {
		log.Info().Str("dir", step.dir).Str("cmd", step.name).Strs("args", step.args).Msg("running")
		if err := runCommand(step.dir, step.name, step.args...); err != nil {
			log.Fatal().Err(err).Str("dir", step.dir).Str("cmd", step.name).Msg("format step failed")
		}
	}

	log.Info().Msg("format completed")
}
