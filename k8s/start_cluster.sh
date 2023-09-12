#!/bin/bash

# Start cluster
minikube start

# Build local Docker images
eval $(minikube -p minikube docker-env)
docker compose build

# Load Kubernetes services
minikube kubectl -- apply -f k8s/enc-server-go-db.yaml
minikube kubectl -- apply -f k8s/enc-server-go-be.yaml
minikube kubectl -- apply -f k8s/enc-server-go-fe.yaml

# Set up port-forwarding
LOCAL_HOST_PORT=7777; REMOTE_PORT=7777
KC_POD_NAME=$(minikube kubectl -- get pods | grep enc-server-go-fe | cut -f1 -d' ')
minikube kubectl --  port-forward $KC_POD_NAME $LOCAL_HOST_PORT:$REMOTE_PORT &
