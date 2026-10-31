package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aktnb/discord-bot-go/internal/bootstrap"
)

// version はボットのバージョン情報を保持します。
// ビルド時に ldflags で上書き可能です: go build -ldflags "-X main.version=v1.0.0"
// デフォルトは "develop" です。
var version = "develop"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := bootstrap.Run(ctx, version); err != nil {
		log.Printf("bot stopped: %v", err)
		os.Exit(1)
	}
}
