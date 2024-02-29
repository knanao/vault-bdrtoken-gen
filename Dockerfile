FROM golang:1.21.7-alpine3.19 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN go build -o /bdrtoken-generator

FROM alpine:3.18
RUN apk --no-cache add ca-certificates

COPY --from=builder /bdrtoken-generator ./
RUN chmod +x ./bdrtoken-generator

ENTRYPOINT ["./bdrtoken-generator"]
