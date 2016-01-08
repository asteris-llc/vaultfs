FROM gliderlabs/alpine:latest
MAINTAINER Brian Hicks <brian@aster.is>

COPY . /go/src/github.com/asteris-llc/vaultfs
RUN apk add --update go git mercurial fuse \
	&& cd /go/src/github.com/asteris-llc/vaultfs \
	&& export GOPATH=/go \
	&& go get \
	&& go build -o /bin/vaultfs \
	&& rm -rf /go \
	&& apk del --purge go git mercurial

RUN mkdir /mounts

ENTRYPOINT [ "/bin/vaultfs" ]
