FROM golang:1.20-alpine as builder
RUN apk add --no-cache --virtual build-deps make git
# Build Elvish
COPY . /go/src/src.elv.sh
RUN make -C /go/src/src.elv.sh get

FROM alpine:3.18
RUN adduser -D elf
RUN apk update && apk add tmux mandoc man-pages vim curl sqlite git
COPY --from=builder /go/bin/elvish /bin/elvish
USER elf
WORKDIR /home/elf
CMD ["/bin/elvish"]
