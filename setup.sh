#!/bin/bash

set -e

usage() {
  echo "Usage: $0 [c|d|a|p]"
  echo "  c: Create kind cluster with Calico CNI and install Ingress-Nginx"
  echo "  d: Delete kind cluster"
  echo "  a: Build app images, load them to cluster, and apply all manifests"
  echo "  p: Port-forward to the Ingress controller for local access (for WSL2)"
  exit 1
}

if [ -z "$1" ]; then
    echo "Error: No argument provided."
    usage
fi

if [ "$1" = "c" ]; then
    echo "--- Creating kind cluster (with Calico CNI) ---"
    kind create cluster --config kind-config.yaml
    
    echo "--- Installing Calico CNI ---"
    kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.28.0/manifests/calico.yaml

    # ★★★ ここからが追加部分 ★★★
    echo "--- Waiting for Calico pods to be ready... ---"
    # Calicoのコントローラー(Deployment)がReadyになるのを待つ
    kubectl wait --for=condition=Ready pod -l k8s-app=calico-kube-controllers -n kube-system --timeout=300s
    # Calicoのノードエージェント(DaemonSet)がReadyになるのを待つ
    kubectl wait --for=condition=Ready pod -l k8s-app=calico-node -n kube-system --timeout=300s
    echo "--- Calico is ready. ---"
    # ★★★ ここまでが追加部分 ★★★

    echo "--- Installing Ingress-Nginx ---"
    kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
    
    echo "--- Waiting for Ingress controller to be ready... ---"
    kubectl wait --namespace ingress-nginx \
      --for=condition=ready pod \
      --selector=app.kubernetes.io/component=controller \
      --timeout=300s
    echo "--- Cluster setup is complete. ---"

elif [ "$1" = "d" ]; then
    echo "--- Deleting kind cluster ---"
    kind delete cluster 

# ... (p オプションと a オプションは変更なし) ...
elif [ "$1" = "p" ]; then
    echo "--- Starting port-forward for Ingress controller ---"
    echo ">>> Access the application at http://localhost:8080 <<<"
    echo ">>> Press Ctrl+C to stop. <<<"
    kubectl port-forward --namespace ingress-nginx service/ingress-nginx-controller 8080:80

elif [ "$1" = "a" ]; then
    echo "--- Building backend image ---"
    docker build -t go-crud-backend:v1 ./backend
    echo "--- Building frontend image ---"
    docker build -t react-crud-frontend:v1 ./frontend
    echo "--- Loading images to kind cluster ---"
    kind load docker-image go-crud-backend:v1
    kind load docker-image react-crud-frontend:v1
    echo "--- Applying Kubernetes manifests in order ---"
    echo "Step 1: Applying Namespace"
    kubectl apply -f k8s/01-namespace.yaml
    echo "Step 2: Applying Secrets"
    kubectl apply -f k8s/02-secrets.yaml
    echo "Step 3: Applying other resources"
    kubectl apply -f k8s/03-mysql.yaml
    kubectl apply -f k8s/04-backend.yaml
    kubectl apply -f k8s/05-frontend.yaml
    kubectl apply -f k8s/06-ingress.yaml
    kubectl apply -f k8s/07-network-policies.yaml
    echo ""
    echo "--- All manifests applied successfully! ---"
    echo ">>> Run './setup.sh p' in a new terminal to access the application. <<<"

else
    echo "Error: Unknown argument '$1'"
    usage
fi
