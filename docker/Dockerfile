# This is the first stage, for building things that will be required by the
# final stage (notably the binary)
FROM golang:1.25.2 AS builder

WORKDIR /go/src/app

# Copy in just the go.mod and go.sum files, and download the dependencies. By
# doing this before copying in the other dependencies, the Docker build cache
# can skip these steps so long as neither of these two files change.
COPY go.mod go.sum ./

# Assuming the source code is collocated to this Dockerfile
# Copy the rest of the source and download dependencies so Docker layer cache
# can be reused when code changes but dependencies don't.
COPY . .

# Download modules first so they are cached in a separate layer.
RUN go mod download

# Build the Go app with CGO disabled and target the main package under cmd/.
# Set GOOS=linux to ensure a Linux binary for the scratch image.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /experia-v10-exporter ./cmd/experia-v10-exporter
# Build a tiny static probe binary we can use in the final scratch image for
# HEALTHCHECK. Keeping this as a Go program means it's statically linked with
# CGO disabled and small.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /probe ./cmd/probe

# Create a "nobody" non-root user for the next image by crafting an /etc/passwd
# file that the next image can copy in. This is necessary since the next image
# is based on scratch, which doesn't have adduser, cat, echo, or even sh.
RUN echo "nobody:x:65534:65534:Nobody:/:" > /etc/passwd

# The second and final stage
FROM scratch

# Copy the binary from the builder stage
COPY --from=builder /experia-v10-exporter /experia-v10-exporter
# Copy probe so the scratch image can run a health check without depending on
# /bin/sh or curl being present.
COPY --from=builder /probe /probe

# Copy the certs from the builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the /etc/passwd file we created in the builder stage. This creates a new
# non-root user as a security best practice.
COPY --from=builder /etc/passwd /etc/passwd

# Run as the new non-root by default
USER nobody

ENTRYPOINT [ "/experia-v10-exporter" ]

# Healthcheck: ensure /metrics is reachable. Use a short timeout so Docker can mark
# the container unhealthy if the app fails to respond quickly. This helps orchestrators
# detect failed containers during startup/restart loops.
HEALTHCHECK --interval=10s --timeout=3s --start-period=30s --retries=3 CMD ["/probe","-url","http://127.0.0.1:9100/metrics","-timeout","2s"]
