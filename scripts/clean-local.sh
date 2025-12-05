#!/bin/bash
set -e

cd "$(dirname "$0")/.."

echo "This will delete all resources in the dfloo-profile namespace"
read -p "Continue? (Y/n): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled"
    exit 0
fi

CURRENT_CONTEXT=$(kubectl config current-context)

if [[ "$CURRENT_CONTEXT" != "docker-desktop" ]]; then
    echo "Not connected to Docker Desktop Kubernetes"
    echo "Current context: $CURRENT_CONTEXT"
    echo ""
    echo "Switch to docker-desktop with:"
    echo "  kubectl config use-context docker-desktop"
    exit 1
fi

echo "Cleaning up local deployment..."

kubectl delete namespace dfloo-profile --ignore-not-found

kubectl delete pv postgres-pv-local --ignore-not-found

echo "Cleaning up local PostgreSQL data..."
rm -rf /tmp/postgres-data

echo "Cleanup complete!"
echo "You can now run a fresh deployment with:"
echo "  ./scripts/deploy-local.sh"