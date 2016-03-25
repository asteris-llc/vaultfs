FROM gliderlabs/alpine:3.3

# build-base
RUN apk add --no-cache build-base

# go
RUN apk add --no-cache go
RUN mkdir /go
ENV GOPATH /go
ENV GO15VENDOREXPERIMENT 1

# glide
RUN apk add --no-cache --virtual=glide-deps curl ca-certificates && \
    mkdir /tmp/glide && \
    curl -L https://github.com/Masterminds/glide/releases/download/0.9.3/glide-0.9.3-linux-amd64.tar.gz | tar -xzv -C /tmp/glide && \
    apk del glide-deps && \
    mv /tmp/glide/**/glide /bin/glide
