package main

import (
	"log/slog"
	"os"

	"github.com/cert-manager/community/security/audit-access/cmd"
)

var version = "dev"

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	if err := cmd.Execute(version, logger); err != nil {
		logger.Error("execution failed", "error", err)
		os.Exit(1)
	}
}
