# --- ステージ1: ビルド環境 ---
FROM node:18-alpine AS builder

WORKDIR /app

# 依存関係をインストール
COPY package*.json ./
RUN npm install

# ソースコードをコピー
COPY . .

# Reactアプリを本番用にビルド
RUN npm run build

# --- ステージ2: 実行環境 ---
FROM nginx:1.25-alpine

# ビルドステージで生成された静的ファイルをNginxの公開ディレクトリにコピー
COPY --from=builder /app/build /usr/share/nginx/html

# Nginxの設定ファイルをコピー
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
