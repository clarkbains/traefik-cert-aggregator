.PHONY: deps all
DOCKER_IMAGE = ghcr.io/clarkbains/cert-agg

all: build/cert-agg

build/cert-agg: deps
	go build -ldflags="-extldflags=-static" -o build/cert-agg cmd/main.go

docker-build:
	docker build . -t $(DOCKER_IMAGE)

docker-push: docker-build
	docker push $(DOCKER_IMAGE)

docker-run: docker-build
	docker run $(DOCKER_IMAGE)

deps:
	go mod download