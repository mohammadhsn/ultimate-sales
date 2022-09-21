SHELL := /bin/bash

VERSION := 1.0

# Go
run:
	go run main.go


# Docker
image:
	docker image build \
		-t github.com/mohammadhsn/ultimate-service \
		-f zarf/docker/Dockerfile \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.


# Kubernetes/Kind
CLUSTER_NAME := ultimate-cluster

kind-up:
	kind create cluster --name $(CLUSTER_NAME) --config zarf/k8s/kind/kind-config.yaml

kind-down:
	kind delete cluster --name $(CLUSTER_NAME)

kind-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces
