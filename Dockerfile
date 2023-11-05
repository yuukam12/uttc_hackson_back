FROM mysql:8.0
ENV LANG ja_JP.UTF-8

# 作業ディレクトリを設定
WORKDIR /Users/matsumotoyuka/uttc_hackson/backend

# ホストのGoプログラムファイルをコンテナにコピー
COPY main.go .
COPY go.mod .
# COPY go.sum .

# Goモジュールをダウンロード
RUN go mod download
# Goプログラムをビルド
RUN go build -o db

# ポートを公開 (必要に応じて)
EXPOSE 8080

# コンテナが起動したときに実行するコマンドを指定
CMD ["./your-app-name"]