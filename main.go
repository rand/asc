package main

import (
	"fmt"
	"os"

	"github.com/rand/asc/cmd"
	"github.com/rand/asc/internal/logger"
)

func main() {
	// Initialize logger
	if err := logger.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to initialize logger: %v\n", err)
	}
	defer logger.Close()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
