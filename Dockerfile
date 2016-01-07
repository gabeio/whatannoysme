FROM golang:alpine

# git/hg required for "go get" & clean apk cache for smallest image size
RUN apk add --update git mercurial && rm -rf /var/cache/apk/*

# if we install these first the build will be cached
RUN go get golang.org/x/crypto/bcrypt
RUN go get gopkg.in/dancannon/gorethink.v1
RUN go get gopkg.in/boj/redistore.v1
RUN go get github.com/labstack/echo

COPY . /go/src/github.com/gabeio/whatannoysme

RUN go install github.com/gabeio/whatannoysme

WORKDIR /go/src/github.com/gabeio/whatannoysme

ENTRYPOINT /go/bin/whatannoysme

# Document that the service listens on port 8080.
EXPOSE 8080
