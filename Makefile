SHELL := /bin/bash

VERSION := 1.0

# Go
go-run:
	go run -ldflags "-X main.build=local" app/services/sales/main.go | go run app/tooling/logfmt/main.go

go-tidy:
	go mod tidy
	go mod vendor

go-expvarmon:
	expvarmon --ports=":4000" --vars="build,requests,goroutines,errors,panics,mem:memstats,Alloc"

go-test:
	go test ./... -count=1
	staticcheck -checks=all ./...

admin:
	go run app/tooling/admin/main.go


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
apply:
	kubectl kustomize zarf/k8s/kind/db-pod | kubectl apply -f -
	kubectl wait --namespace=database-system --timeout=120s --for=condition=Available deployment/database-pod
	kubectl kustomize zarf/k8s/kind/sales-pod | kubectl apply -f -

k8s-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch

sales-restart:
	kubectl rollout restart deployment sales-pod

sales-update: docker-build kind-load sales-restart

sales-update-apply: docker-build kind-load apply

sales-status:
	kubectl get pods -o wide --watch

sales-logs:
	kubectl logs \
		-l app=sales \
		--all-containers=true \
		--tail 100 -f | go run app/tooling/logfmt/main.go | go run app/tooling/logfmt/main.go

sales-describe:
	kubectl describe pod -l app=sales

db-status:
	kubectl get pod -o wide --watch --namespace=database-system
