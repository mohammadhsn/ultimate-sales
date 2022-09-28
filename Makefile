SHELL := /bin/bash

VERSION := 1.0

# Go
go-run:
	go run -ldflags "-X main.build=local" main.go


# Docker
IMAGE := sales-service:$(VERSION)

docker-build:
	docker image build \
		-t $(IMAGE) \
		-f zarf/docker/Dockerfile \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.


# Kind
kind-up:
	kind create cluster --config zarf/k8s/kind/kind-config.yaml
	kubectl config set-context --current --namespace=sales-system

kind-down:
	kind delete cluster

kind-load:
	kind load docker-image $(IMAGE)

# Kubernetes
k8s-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch

sales-restart:
	kubectl rollout restart deployment sales-pod

sales-apply:
	cat zarf/k8s/base/sales-pod.yaml | kubectl apply -f -

sales-status:
	kubectl get pods -o wide --watch

sales-logs:
	kubectl logs \
		-l app=sales \
		--all-containers=true \
		--tail 100 -f
