# ビルドステージ
FROM golang:1.24-alpine AS builder
WORKDIR /app

# golang-migrate CLI をビルド（postgres タグ）
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.version=${VERSION}" -o bot ./cmd/bot

# 実行ステージ
FROM alpine:3
# ca-certificates, tzdata: 基本パッケージ
# ffmpeg: ラジオストリームのデコード・エンコードに使用
RUN apk --no-cache add ca-certificates tzdata ffmpeg
WORKDIR /app
COPY --from=builder /app/bot .
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY db/migrations ./migrations
COPY entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh
ENTRYPOINT ["./entrypoint.sh"]
