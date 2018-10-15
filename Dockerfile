FROM golang:alpine as builder
RUN apk update && \
    apk add --virtual build-deps make git
COPY . /go/src/github.com/elves/elvish
RUN make -C /go/src/github.com/elves/elvish get

FROM alpine
COPY --from=builder /go/bin/elvish /bin/elvish
RUN adduser -D elf
RUN apk update && apk add tmux man man-pages vim curl
USER elf
WORKDIR /home/elf
CMD ["/bin/elvish"]
