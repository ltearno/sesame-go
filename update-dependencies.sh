#!/bin/bash
export GOPATH=$(pwd)

go env

go get -u github.com/jteeuwen/go-bindata/...
go get github.com/julienschmidt/httprouter
go get -u github.com/gbrlsnchs/jwt
go get github.com/google/uuid