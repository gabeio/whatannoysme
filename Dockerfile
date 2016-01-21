FROM golang:alpine

RUN apk add --update git mercurial && \
	rm -rf /var/cache/apk/*

EXPOSE 8080

COPY . /go/src/github.com/gabeio/whatannoysme

WORKDIR /go/src/github.com/gabeio/whatannoysme

RUN go get github.com/skelterjohn/wgo && \
	wgo restore && \
	wgo install whatannoysme && \
	cp bin/whatannoysme /go/bin

CMD ["/go/bin/whatannoysme"]
