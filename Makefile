SHELL := /bin/bash

VERSION := $(sh git rev-parse --short HEAD)

# Go
go-run:
	go run -ldflags "-X main.build=local" main.go


# Docker
IMAGE := sales-service:latest

docker-build:
	docker image build \
		-t $(IMAGE) \
		-f zarf/docker/Dockerfile \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.


# Kind
CLUSTER := ultimate-cluster

kind-up:
	kind create cluster \
		--name $(CLUSTER) \
		--image kindest/node:v1.17.0@sha256:9512edae126da271b66b990b6fff768fbb7cd786c7d39e86bdf55906352fdf62 \
		--config zarf/k8s/kind/kind-config.yaml
	kubectl config set-context --current --namespace=sales-system

kind-down:
	kind delete cluster --name $(CLUSTER)

kind-load:
	kind load docker-image $(IMAGE) --name $(CLUSTER)

# Kubernetes
k8s-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch

sales-restart:
	kubectl rollout restart deployment sales-pod

sales-apply:
	cat zarf/k8s/base/sales-pod.yaml | kubectl apply -f -
	kustomize build zarf/k8s/kind/kustomization.yaml | kubectl apply -f -

sales-status:
	kubectl get pods -o wide --watch

sales-logs:
	kubectl logs \
		-l app=sales \
		--all-containers=true \
		--tail 100 -f
