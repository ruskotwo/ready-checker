TAG=github.com/ruskotwo/ready-checker:latest
DOCKER_BUILD_OPTIONS ?=--platform linux/amd64

wire:
	cd cmd/factory && wire ; cd ../..

golang_build:
	docker build $(DOCKER_BUILD_OPTIONS) \
		-t $(TAG) -f ./docker/golang.Dockerfile .
