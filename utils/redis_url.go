package utils

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func GetRedisURL() string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading .env file")
	}
	// Redis
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		panic("REDIS_URL not set")
	}
	return redisURL
}
