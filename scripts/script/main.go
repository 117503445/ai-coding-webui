package main

import (
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
}

var dirProjectRoot = mustProjectRoot()

func main() {
	ctx := kong.Parse(&cli)
	if err := ctx.Run(); err != nil {
		log.Fatal().Err(err).Msg("run failed")
	}
}
