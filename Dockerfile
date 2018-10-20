FROM golang:alpine as builder
RUN apk update && \
    apk add --virtual build-deps make git
# Build gotty
RUN go get -d github.com/yudai/gotty && \
    git -C /go/src/github.com/yudai/gotty checkout release-1.0 && \
    go get github.com/yudai/gotty
# Build Elvish
COPY . /go/src/github.com/elves/elvish
RUN make -C /go/src/github.com/elves/elvish get

FROM alpine
COPY --from=builder /go/bin/elvish /bin/elvish
COPY --from=builder /go/bin/gotty /bin/gotty
RUN adduser -D elf
RUN apk update && apk add tmux man man-pages vim curl
USER elf
WORKDIR /home/elf
EXPOSE 8080
CMD ["/bin/elvish"]
