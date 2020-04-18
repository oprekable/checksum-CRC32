FROM golang AS builder

COPY . /go/src/github.com/oprekable/checksum-CRC32/
WORKDIR /go/src/github.com/oprekable/checksum-CRC32/

RUN make quicktest
RUN \
  CGO_ENABLED=0 \
  make
RUN ./checksum-CRC32

# Begin final image
FROM alpine:latest

RUN apk --no-cache add ca-certificates fuse

COPY --from=builder /go/src/github.com/oprekable/checksum-CRC32/checksum-CRC32 /usr/local/bin/

ENTRYPOINT [ "checksum-CRC32" ]

WORKDIR /data