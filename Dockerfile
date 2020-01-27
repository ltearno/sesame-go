FROM golang:1.12-alpine

RUN apk add make git

RUN mkdir /.cache && chmod ugo+rw /.cache

ADD update-dependencies.sh ./
RUN ls
RUN sh ./update-dependencies.sh

ADD Makefile ./
ADD src ./src/
ADD assets ./assets/
ADD configuration.json ./

RUN make build-embed-assets
RUN make install

ENTRYPOINT [ "/go/bin/${APP_NAME}" ]
