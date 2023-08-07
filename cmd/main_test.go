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

func setupLogger(t *testing.T) {
	t.Helper()

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

func TestFileGraphSetup(t *testing.T) {
	setupLogger(t)
	path := "../assets/out.json"

	g := *streets.NewGraphBuilder().FromJsonFile(path).Build()

	if g == nil {
		t.Errorf("Graph is nil")
	}

	size, err := g.Size()
	if err != nil {
		t.Errorf("Error getting graph size: %s", err)
	}

	if size < 10 {
		t.Errorf("Graph size is too small: %d", size)
	}
}
