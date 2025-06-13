# ビルドステージ
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Go modulesをコピーして依存関係をダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# バイナリをビルド（Cloud Run用サーバー）
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# 実行ステージ
FROM alpine:latest

# CA証明書を追加（HTTPS通信のため）
RUN apk --no-cache add ca-certificates
WORKDIR /root/

# ビルドしたバイナリをコピー
COPY --from=builder /app/server .

# ポート8080を開放
EXPOSE 8080

# アプリケーションを実行
CMD ["./server"]