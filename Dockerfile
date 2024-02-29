FROM golang:1.22-alpine3.19 as builder
RUN apk add --no-cache --virtual build-deps make git
# Build Elvish
COPY . /go/src/src.elv.sh
RUN make -C /go/src/src.elv.sh get

FROM alpine:3.19
COPY --from=builder /go/bin/elvish /bin/elvish
CMD ["/bin/elvish"]
