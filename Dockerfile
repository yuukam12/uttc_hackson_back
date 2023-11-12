FROM mysql:8.0
ENV LANG ja_JP.UTF-8

# 作業ディレクトリを設定
WORKDIR /app

# ホストのGoプログラムファイルをコンテナにコピー
COPY main.go .
COPY go.mod .
# COPY go.sum .

# Goプログラムをビルド
RUN go build -o main.go

# ポートを公開 (必要に応じて)
EXPOSE 3306

# コンテナが起動したときに実行するコマンドを指定
CMD ["./main"]