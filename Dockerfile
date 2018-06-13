FROM golang:alpine

RUN apk update && \
    apk add --virtual build-deps make git

COPY . /go/src/github.com/elves/elvish
RUN make -C /go/src/github.com/elves/elvish get

RUN apk del --purge build-deps

WORKDIR /root
CMD /go/bin/elvish
