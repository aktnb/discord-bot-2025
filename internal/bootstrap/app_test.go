package bootstrap

import (
	"context"
	"strings"
	"testing"

	"github.com/aktnb/discord-bot-go/internal/config"
)

func TestRunConfigurationError(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("DISCORD_TOKEN", "")
	t.Setenv("DATABASE_URL", "")
	err := Run(context.Background(), "test-version")
	if err == nil || !strings.Contains(err.Error(), "failed to load configuration") {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestRunInvalidDatabaseConfiguration(t *testing.T) {
	err := run(context.Background(), config.Config{
		DiscordToken: "test-token",
		DatabaseURL:  "postgres://localhost:invalid/database",
	}, "test-version")
	if err == nil || !strings.Contains(err.Error(), "failed to create database connection pool") {
		t.Fatalf("run() error = %v", err)
	}
}
