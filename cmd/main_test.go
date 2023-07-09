package main

import (
	"os"
	"testing"

	"pchpc/utils"

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

func setupDB(b *testing.B) {
	b.Helper()

	utils.SetDBPath("../assets/db.sqlite")
}

func BenchmarkRunWithoutRoutines(b *testing.B) {
	setupLogger(b)
	setupDB(b)
	routines := false
	g := streets.NewGraph(utils.GetDbPath())

	run(&g, &b.N, &minSpeed, &maxSpeed, &routines)
}

func BenchmarkRunRoutines(b *testing.B) {
	setupLogger(b)
	setupDB(b)
	routines := true

	g := streets.NewGraph(utils.GetDbPath())
	run(&g, &b.N, &minSpeed, &maxSpeed, &routines)
}
