# Build the manager binary
FROM --platform=$BUILDPLATFORM golang:1.20 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
COPY vendor/ vendor/


COPY main.go main.go
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

FROM --platform=$TARGETPLATFORM alpine
WORKDIR /controller
COPY --from=builder /workspace/manager .
USER root:root

ENTRYPOINT ["/controller/manager"]
