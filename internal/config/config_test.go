package config

import "testing"

func TestLoad(t *testing.T) {
	for _, tt := range []struct {
		name, token, database string
		wantErr               bool
	}{
		{"valid", "token", "postgres://example", false},
		{"missing token", "", "postgres://example", true},
		{"missing database", "token", "", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			t.Setenv("DISCORD_TOKEN", tt.token)
			t.Setenv("DATABASE_URL", tt.database)
			cfg, err := Load()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Load error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && (cfg.DiscordToken != tt.token || cfg.DatabaseURL != tt.database) {
				t.Fatalf("unexpected config: %+v", cfg)
			}
		})
	}
}
