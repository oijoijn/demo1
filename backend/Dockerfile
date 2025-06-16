# --- ステージ1: ビルド環境 ---
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 依存関係をキャッシュ
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# アプリケーションをビルド
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/main .

# --- ステージ2: 実行環境 ---
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# ビルドステージから実行可能ファイルのみをコピー
COPY --from=builder /app/main .

# 8080ポートを公開
EXPOSE 8080

# コンテナ起動時に実行するコマンド
CMD ["./main"]
