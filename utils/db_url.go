package utils

import "os"

var dbPath string

// SetDBPath sets the path to the database file
func SetDBPath(path string) {
	dbPath = path
}

// GetDbPath returns the path to the database file
func GetDbPath() string {
	//err := godotenv.Load()
	//if err != nil {
	//	log.Warn().Err(err).Msg("Error loading .env file")
	//}
	//// SQLITE
	//dbPath := os.Getenv("DB_PATH")

	if dbPath == "" {
		panic("DB_PATH not set")
	}

	// Check if file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		panic("DB_PATH / database file does not exist")
	}

	return dbPath
}
