package utils

import "github.com/rs/zerolog/log"

// IfErr is a helper function that logs an error if it's not nil.
func IfErr(err error) error {
	if err != nil {
		log.Error().Err(err).Msg("Error occurred.")
		return err
	}
	return nil
}
