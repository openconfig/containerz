FROM golang:1.25-alpine AS builder

# Run docker build from root of clio.
RUN mkdir -p /go/src/github.com/openconfig/containerz
COPY . /go/src/github.com/openconfig/containerz
WORKDIR /go/src/github.com/openconfig/containerz
RUN GOOS=linux go build -C . -o containerz

# Run second stage for the container that we actually run.
FROM alpine:latest
RUN mkdir /app
COPY --from=builder go/src/github.com/openconfig/containerz/containerz /app
CMD ["/app/containerz", "start"]
