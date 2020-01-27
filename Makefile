export GOPATH := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
export APP_NAME := $(shell basename $(GOPATH))

COMMIT := latest

all: install

.PHONY: build-prepare
build-prepare:
	@echo "updating dependencies..."
	@./update-dependencies.sh

.PHONY: build-embed-assets
build-embed-assets:
	@echo "embedding assets..."
	@./bin/go-bindata -o src/application/assetsgen/assets.go -pkg assetsgen assets/...

.PHONY: build
build: build-embed-assets
	@echo "build binaries..."
	@./build-releases.sh

.PHONY: install
install: build-embed-assets
	@echo "install binaries..."
	#@go install $(APP_NAME)

.PHONY: run-serve
run-serve:
	./$(APP_NAME) serve

.PHONY: build-docker
build-docker:
	cat Dockerfile | sed 's/$${APP_NAME}/${APP_NAME}/g' | docker build . -f - -t $(APP_NAME):${COMMIT}

.PHONY: run-serve-docker
run-docker:
	docker run -it --rm -p 8080:8080 --user $(shell id -u):$(shell id -g) $(APP_NAME):${COMMIT}

.PHONY: run-serve-interactive
run-docker-interactive:
	docker run -it --rm -p 8080:8080 -v "$(shell pwd)":/localhost --entrypoint sh $(APP_NAME):${COMMIT}
