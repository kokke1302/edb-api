# ビルドステージ
FROM golang:1.27-alpine AS builder

WORKDIR /app

# COPY go.mod ./
# COPY go.sum ./ # パッケージ追加時は有効化してください

COPY . .

# Goのプログラムをビルド（コンパイル）し、server という名前の実行ファイルを作る。
# CGO_ENABLED=0: C言語のライブラリに頼らない（純粋なGoだけで動く）ファイルにする。
# GOOS=linux: Linux上で動く形式でビルドする。
RUN CGO_ENABLED=0 GOOS=linux go build -o server .
RUN chmod +x server

# 実行ステージ（軽量コンテナ）
FROM alpine:latest

# アプリが外部とHTTPS通信（SSL/TLS）を行うために必要な証明書のパッケージをインストール。
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 最初のステージ（builder）で作った server という実行ファイルだけを、この軽量なコンテナの中にコピーしてくる。
COPY --from=builder /app/server .

EXPOSE 8080

# コンテナが起動したときに、自動的に ./serverを実行する。
CMD ["./server"]