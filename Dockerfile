# syntax = docker/dockerfile:1.5.2

FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder
ARG BUILDKIT_SBOM_SCAN_STAGE=true
WORKDIR /build
ARG TARGETOS TARGETARCH
ENV GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg --mount=source=.,target=. \
    go build -ldflags="-s -w" -o /frontend/microb ./cmd/microb/main.go


FROM scratch

WORKDIR /home/nonroot
COPY --link --from=builder --chown=65532:65532 --chmod=500 /frontend/microb /home/nonroot/microb

ENTRYPOINT ["/home/nonroot/microb"]
