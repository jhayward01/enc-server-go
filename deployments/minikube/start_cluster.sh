#!/bin/bash

# Start cluster
minikube start

# Build local Docker images
eval $(minikube -p minikube docker-env)
docker compose build

# Load Kubernetes services and deployments
# TODO think these paths are wrong
minikube kubectl -- apply -f deployments/minikube/enc-server-go-db.yaml
minikube kubectl -- apply -f deployments/minikube/enc-server-go-be.yaml
minikube kubectl -- apply -f deployments/minikube/enc-server-go-fe.yaml
