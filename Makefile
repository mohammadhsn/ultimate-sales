SHELL := /bin/bash

VERSION := 1.0

# Go
go-run:
	go run -ldflags "-X main.build=local" app/services/sales/main.go

go-tidy:
	go mod tidy
	go mod vendor


# Docker

docker-build:
	docker image build \
		-t sales-image:$(VERSION) \
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
	#cd zarf/k8s/kind; kustomize edit set image sales-image=sales-amd64:$(VERSION)
	kind load docker-image sales-image:$(VERSION)

# Kubernetes
k8s-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch

sales-restart:
	kubectl rollout restart deployment sales-pod

sales-apply:
	kubectl kustomize zarf/k8s/kind | kubectl apply -f -

sales-update: docker-build kind-load sales-restart

sales-update-apply: docker-build kind-load sales-apply

sales-status:
	kubectl get pods -o wide --watch

sales-logs:
	kubectl logs \
		-l app=sales \
		--all-containers=true \
		--tail 100 -f

sales-describe:
	kubectl describe pod -l app=sales
