package main

import (
	"os"
	"testing"

	"pchpc/streets"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	minSpeed = 5.5
	maxSpeed = 8.5
)

func setupLogger(b *testing.B) {
	b.Helper()

	// Logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	runLogFile, _ := os.OpenFile(
		"main.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0o664,
	)
	multi := zerolog.MultiLevelWriter(os.Stdout, runLogFile)
	log.Logger = zerolog.New(multi).With().Timestamp().Logger()
}

func BenchmarkRunWithoutRoutines(b *testing.B) {
	setupLogger(b)
	routines := false
	g := streets.NewGraph()

	run(&g, &b.N, &minSpeed, &maxSpeed, &routines)
}

func BenchmarkRunRoutines(b *testing.B) {
	setupLogger(b)
	routines := true

	g := streets.NewGraph()
	run(&g, &b.N, &minSpeed, &maxSpeed, &routines)
}
