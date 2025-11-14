FROM golang:1.25-alpine AS builder

# Run docker build from root of clio.
RUN mkdir -p /go/src/github.com/openconfig/containerz
RUN GOOS=linux go build -C . -o containerz

HEALTHCHECK --interval=30s --timeout=3s --retries=3 CMD curl -f http://localhost/health || exit 1

# Run second stage for the container that we actually run.
FROM alpine:3.22
RUN useradd -m appuser
USER appuser

RUN mkdir /app
COPY --from=builder go/src/github.com/openconfig/containerz/containerz /app
CMD ["/app/containerz", "start"]