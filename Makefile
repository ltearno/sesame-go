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
	@go install application

.PHONY: run-serve
run-serve:
	./$(APP_NAME) serve

.PHONY: build-docker
build-docker: tls.key.pem
	docker build . -t $(APP_NAME):${COMMIT}

.PHONY: run-serve-docker
run-docker:
	docker run -it --rm -p 8443:8443 --user $(shell id -u):$(shell id -g) $(APP_NAME):${COMMIT} serve

.PHONY: run-serve-interactive
run-docker-interactive:
	docker run -it --rm -p 8443:8443 -v "$(shell pwd)":/localhost --entrypoint sh $(APP_NAME):${COMMIT}

tls.key.pem:
	@echo creating TLS certificates...
	@openssl req -x509 -nodes -newkey rsa:2048 -subj "/C=FR/ST=France/L=South/O=LTE Consulting/OU=Factory/CN=localhost" -keyout tls.key.pem -out tls.cert.pem -days 3650