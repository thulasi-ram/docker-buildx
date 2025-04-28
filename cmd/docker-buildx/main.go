package main

import (
	"os"

	"codeberg.org/woodpecker-plugins/plugin-docker-buildx/plugin"
	"github.com/joho/godotenv"
)

var Version = "unknown"

func main() {
	if _, err := os.Stat("/run/drone/env"); err == nil {
		_ = godotenv.Overload("/run/drone/env")
	}

	if envFile, set := os.LookupEnv("PLUGIN_ENV_FILE"); set {
		_ = godotenv.Overload(envFile)
	}

	plugin.New(nil, Version).Run()
}
