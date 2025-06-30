FROM golang:1.24-alpine3.22 AS builder

RUN apk add --no-cache git

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY . .

# Build
RUN CGO_ENABLED=0 go build -o webhook -ldflags '-w -extldflags "-static"' .

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

COPY --from=builder /workspace/webhook /usr/local/bin/webhook

ENTRYPOINT ["webhook"]
