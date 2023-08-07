package main

import (
	"io"
	"os"
	"testing"

	"pchpc/streets"

	"pchpc/utils"

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

func setupDB(t *testing.T) {
	t.Helper()

	utils.SetDBPath("../assets/db.sqlite")
}

func TestFileGraphSetup(t *testing.T) {
	setupLogger(t)

	jsonFile, err := os.Open("../assets/out.json")
	if err != nil {
		t.Fatal(err)
	}

	defer jsonFile.Close()

	// read our opened jsonFile as a byte array.
	byteValue, err := io.ReadAll(jsonFile)

	g, err := streets.NewGraphFromJSON(byteValue)
	if err != nil {
		t.Fatal(err)
	}

	if g == nil {
		t.Fatal("Graph is nil")
	}
}
