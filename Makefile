export GOPATH := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
export APP_NAME := $(shell basename $(GOPATH))

COMMIT := latest

all: install

.PHONY: build-prepare
build-prepare:
	@echo "updating dependencies..."
	@go get -u github.com/jteeuwen/go-bindata/...
	@go get github.com/julienschmidt/httprouter
	@go get -u github.com/gbrlsnchs/jwt

.PHONY: build-embed-assets
build-embed-assets:
	@echo "embedding assets..."
	@./bin/go-bindata -o assetsgen/assets.go -pkg assetsgen assets/...

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
	cat Dockerfile \
		| sed 's/$${APP_NAME}/${APP_NAME}/g' \
		| docker build . -f - -t $(APP_NAME):${COMMIT}

.PHONY: run-serve-docker
run-serve-docker:
	docker run -it --rm -p 9988:9988 --user $(shell id -u):$(shell id -g) $(APP_NAME):${COMMIT} -port 9988 serve /test-dir

.PHONY: run-serve-interactive
run-docker-interactive:
	docker run -it --rm -v "$(shell pwd)":/localhost --user $(shell id -u):$(shell id -g) --entrypoint sh $(APP_NAME):${COMMIT}
