SHELL := /bin/bash

VERSION := 1.0

run:
	go run main.go

image:
	docker image build \
		-t github.com/mohammadhsn/ultimate-service \
		-f zarf/docker/Dockerfile \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.
