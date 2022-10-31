package main

import "os"

var databaseContainer DatabaseContainer

func main() {
	databaseContainer = createDatabaseContainer()
	runApiServer()
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
