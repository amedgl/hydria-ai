// main.go
// HydrIA AI — entry point.
// All logic lives in internal/cmd. This file stays intentionally minimal.
package main

import (
	"github.com/hydria-ai/hydria/internal/cmd"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load() //nolint:errcheck — .env is optional
	cmd.Execute()
}
