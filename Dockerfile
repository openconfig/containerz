FROM golang:1.25-alpine AS builder

# Run docker build from root of clio.
RUN mkdir -p /go/src/github.com/openconfig/containerz
COPY . /go/src/github.com/openconfig/containerz
WORKDIR /go/src/github.com/openconfig/containerz
RUN GOOS=linux go build -C . -o containerz

# Run second stage for the container that we actually run.
FROM alpine:3.22

# Install curl, which is commonly used for health checks.
RUN apk add --no-cache curl=8.6.0-r0 && \
    addgroup -S appgroup && adduser -S appuser -G appgroup && \
    mkdir /app && \
    chown -R appuser:appgroup /app

# ADD HEALTHCHECK: Assumes your Go app serves a health endpoint on port 8080.
# You MUST change the port/path if your application uses a different one.
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
  CMD curl --fail http://localhost:8080/health || exit 1

# Switch to the non-root user for security.
USER appuser

# Copy the built Go binary from the builder stage
COPY --from=builder go/src/github.com/openconfig/containerz/containerz /app
CMD ["/app/containerz", "start"]
