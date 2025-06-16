# !/bin/bash

if [ $1 = "c" ]; then
    echo 'create kind cluster'
    kind create cluster --config kind-config.yaml
elif [ $1 = "d" ]; then
    echo 'delete kind cluster'
    kind delete cluster 
elif [ $1 = "p" ]; then
    echo 'port-forward'
    kubectl port-forward service/web -n app-demo 8080:80
elif [ $1 = "a" ]; then
    echo 'kubectl apply'
    kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
    docker build -t go-crud-backend:v1 ./backend
    docker build -t react-crud-frontend:v1 ./frontend
    kind load docker-image go-crud-backend:v1
    kind load docker-image react-crud-frontend:v1
    # まずはSecretを先に適用
    kubectl apply -f k8s/02-secrets.yaml

    # 残りのマニフェストを適用
    kubectl apply -f k8s/
else
    echo 'try again'
fi
