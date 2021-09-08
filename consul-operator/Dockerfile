# Build the manager binary
FROM golang:1.16 as builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY libs/ libs/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o consul-operator main.go

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ENV RESREQ_DIR=/usr/src/app/resource-reqs-generated \
    DEPLOYMENT_DIR=/usr/src/app

WORKDIR /
COPY --from=builder /workspace/consul-operator .

COPY deployment/helm /usr/local/bin/helm
RUN chmod +x  /usr/local/bin/helm
COPY deployment/app-deployment /usr/src/app/app-deployment
COPY deployment/resource-reqs /usr/src/app/resource-reqs
RUN chmod -R 770 /usr/src/app/

USER 65532:65532

ENTRYPOINT ["/consul-operator"]
