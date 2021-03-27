FROM golang:1.16-alpine as builder
RUN apk update && \
    apk add --virtual build-deps make git
# Build Elvish
COPY . /go/src/src.elv.sh
RUN make -C /go/src/src.elv.sh get

FROM alpine:3.13
COPY --from=builder /go/bin/elvish /bin/elvish
RUN adduser -D elf
RUN apk update && apk add tmux mandoc man-pages vim curl git
USER elf
WORKDIR /home/elf
CMD ["/bin/elvish"]
