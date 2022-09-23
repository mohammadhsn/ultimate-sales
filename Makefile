SHELL := /bin/bash

VERSION := 1.0

# Go
run:
	go run main.go


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
	kind create cluster --name $(CLUSTER) --config zarf/k8s/kind/kind-config.yaml

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
	kubectl rollout restart deployment sales-pod --namespace sales-system

sales-apply:
	cat zarf/k8s/base/sales-pod.yaml | kubectl apply -f -


sales-status:
	kubectl get pods -o wide --watch --namespace sales-system

sales-logs:
	kubectl logs \
		-l app=sales \
		--all-containers=true \
		--namespace sales-system \
		--tail 100 -f
