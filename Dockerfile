FROM golang:wheezy

# if we install these first the build will be cached
RUN go get github.com/zenazn/goji
RUN go get golang.org/x/crypto/bcrypt
RUN go get gopkg.in/mgo.v2
RUN go get gopkg.in/boj/redistore.v1

ADD . /go/src/github.com/gabeio/whatannoysme

RUN go install github.com/gabeio/whatannoysme

RUN cd /go/src/github.com/gabeio/whatannoysme
ENTRYPOINT /go/bin/whatannoysme

# Document that the service listens on port 8080.
EXPOSE 8080
