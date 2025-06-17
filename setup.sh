#!/usr/bin/env bash
set -euo pipefail

# ===== 基本設定 =====
CLUSTER_NAME="crud-cluster"
CONFIG_FILE="kind-config.yaml"
MANIFEST_DIR="k8s"            # マニフェスト格納ディレクトリ

# 適用順（Namespace → Secret → …）
MANIFESTS=(
  # "$MANIFEST_DIR/00-label-ingress-nginx.yaml" 
  "$MANIFEST_DIR/01-namespace.yaml"
  "$MANIFEST_DIR/02-secrets.yaml"
  "$MANIFEST_DIR/03-mysql.yaml"
  "$MANIFEST_DIR/04-backend.yaml"
  "$MANIFEST_DIR/05-frontend.yaml"
  "$MANIFEST_DIR/06-ingress.yaml"
  "$MANIFEST_DIR/07-network-policies.yaml"
)

# Kind ノードへロードするローカルイメージ
IMAGES=(
  go-crud-backend:v1
  react-crud-frontend:v1
)

# ===== 関数 =====
load_images() {
  for img in "${IMAGES[@]}"; do
    if [[ -n $(docker images -q "$img") ]]; then
      echo "Loading $img into kind..."
      kind load docker-image "$img" --name "$CLUSTER_NAME"
    else
      echo "⚠️  Local image $img not found. Build it first." >&2
    fi
  done
}

label_nodes_for_ingress() {
  echo "Labeling only the control-plane node with ingress-ready=true ..."

  # control-plane ラベルが付いたノードを 1 件取得
  ctrl=$(kubectl get nodes \
          -l node-role.kubernetes.io/control-plane= \
          -o jsonpath='{.items[0].metadata.name}')

  if [[ -z "$ctrl" ]]; then
    echo "❌ control-plane node not found; aborting" >&2
    exit 1
  fi

  # 既存ラベルをクリアして control-plane のみに付与
  kubectl label nodes --all ingress-ready- || true
  kubectl label node "$ctrl" ingress-ready=true --overwrite
}

install_ingress() {
  echo "Installing ingress-nginx controller ..."
  kubectl apply -f \
    https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

  # Controller Pod が Ready になるまで
  echo "⏳ Waiting for ingress-nginx controller Pod to be Ready ..."
  kubectl wait -n ingress-nginx \
    --for=condition=ready pod \
    --selector=app.kubernetes.io/component=controller \
    --timeout=180s

  # Admission Webhook Service が Endpoints を持つまで
  echo "⏳ Waiting for admission webhook service endpoints ..."
  for _ in {1..60}; do
    if [[ "$(kubectl -n ingress-nginx \
           get ep ingress-nginx-controller-admission \
           -o jsonpath='{.subsets}' 2>/dev/null)" != "" ]]; then
      echo "✅ admission webhook service is ready"
      break
    fi
    sleep 2
  done

  # まだ空なら警告だけ出して続行
  if [[ "$(kubectl -n ingress-nginx \
        get ep ingress-nginx-controller-admission \
        -o jsonpath='{.subsets}' 2>/dev/null)" == "" ]]; then
    echo "⚠️  admission webhook endpoints are still empty; continuing anyway"
  fi
}

apply_manifests() {
  for f in "${MANIFESTS[@]}"; do
    echo "Applying $f ..."
    kubectl apply -f "$f"
  done
}

usage() {
  cat <<EOF
Usage: $0 {up|apply|reload-images|down}

 up             Kind クラスタを作成し、イメージをロードして Ingress Controller をインストール
 apply          k8s/01-*.yaml ～ 07-*.yaml をクラスタに適用
 reload-images  ローカルイメージを Kind ノードに再ロード
 down           Kind クラスタを削除
EOF
  exit 1
}

# ===== メイン =====
case "${1:-}" in
  up)
    kind create cluster --name "$CLUSTER_NAME" --config "$CONFIG_FILE"
    load_images
    label_nodes_for_ingress
    install_ingress
    ;;
  apply)
    apply_manifests
    ;;
  reload-images)
    load_images
    ;;
  down)
    kind delete cluster --name "$CLUSTER_NAME"
    ;;
  *)
    usage
    ;;
esac
