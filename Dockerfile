FROM golang:1.16-alpine AS builder

WORKDIR /src

COPY . .

ENV CGO_ENABLED=0

RUN go build -trimpath

FROM alpine AS bin

COPY --from=builder /src/haniel /usr/bin/haniel

ENTRYPOINT ["/usr/bin/haniel"]

