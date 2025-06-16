# クラスタ作成（イメージロード & Ingress 完全起動まで待機）
./setup.sh up

# マニフェスト適用
./setup.sh apply

# コードを修正してイメージだけ再ロード
docker build -t go-crud-backend:v1 ./backend
./setup.sh reload-images

# クラスタ削除
./setup.sh down
