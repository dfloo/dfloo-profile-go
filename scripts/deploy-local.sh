#!/bin/bash
set -e

cd "$(dirname "$0")/.."

echo "Deploying to local Docker Desktop Kubernetes cluster..."

CURRENT_CONTEXT=$(kubectl config current-context)

if [[ "$CURRENT_CONTEXT" != "docker-desktop" ]]; then
    echo "Not connected to Docker Desktop Kubernetes"
    echo "Current context: $CURRENT_CONTEXT"
    echo ""
    echo "Switch to docker-desktop with:"
    echo "  kubectl config use-context docker-desktop"
    exit 1
fi

if ! kubectl cluster-info &>/dev/null; then
    echo "Kubernetes cluster is not responding"
    exit 1
fi

echo "Connected to Docker Desktop Kubernetes."

echo "Building docker image..."
docker build -t dfloo-profile-go:latest .

echo "Applying k8s manifests..."
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/local/configmap.yaml
kubectl apply -f k8s/local/secrets.yaml
kubectl apply -f k8s/local/postgres-pvc.yaml
kubectl apply -f k8s/local/postgres.yaml

echo "Waiting for PostgreSQL..."
kubectl wait --for=condition=ready pod -l app=postgres -n dfloo-profile --timeout=300s

echo "Creating migrations ConfigMap..."
kubectl create configmap db-migrations \
  --from-file=db/migrations/ \
  --namespace=dfloo-profile \
  --dry-run=client -o yaml | kubectl apply -f -

echo "Running Migrations..."

kubectl delete job db-migration -n dfloo-profile --ignore-not-found
kubectl apply -f k8s/local/migration-job.yaml
kubectl wait --for=condition=complete job/db-migration -n dfloo-profile --timeout=300s

echo "Deploying API..."
kubectl apply -f k8s/local/api.yaml

echo "Restarting API deployment to use new image..."
kubectl rollout restart deployment/dfloo-profile-api -n dfloo-profile

echo "Waiting for API to be ready..."
kubectl rollout status deployment/dfloo-profile-api -n dfloo-profile --timeout=300s

echo "Deployment complete."
echo "API available at: http://localhost:30080"
echo "To check status: kubectl get all -n dfloo-profile"
echo "To view logs: kubectl logs -f deployment/dfloo-profile-api -n dfloo-profile"
