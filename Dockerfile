FROM golang:1.12-alpine

RUN apk add make git

RUN mkdir /.cache && chmod ugo+rw /.cache

WORKDIR /sesame-go

ADD update-dependencies.sh ./
RUN sh ./update-dependencies.sh

ADD Makefile ./
ADD src ./src/
ADD assets ./assets/
ADD configuration.json ./

RUN make build-embed-assets
RUN make install

ADD tls.key.pem ./
ADD tls.cert.pem ./

RUN chown -R 1000:1000 /sesame-go

ENTRYPOINT [ "/sesame-go/bin/application" ]
