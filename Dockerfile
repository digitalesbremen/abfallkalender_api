# The builder always runs natively on the build platform and cross-compiles for
# the target. Go does this out of the box, so no QEMU emulation is involved --
# which is what made the previous multi-arch build slow.
FROM --platform=$BUILDPLATFORM golang:1.26 AS builder
LABEL stage=intermediate

WORKDIR /app

# Dependencies first: this layer stays cached across pure code changes.
COPY go.mod go.sum ./
RUN go mod download

COPY main.go open-api-3.yaml ./
COPY src/backend ./src/backend

# Substitute the release version into the OpenAPI spec before compiling; the
# spec is embedded into the binary via //go:embed in main.go.
ARG VERSION=dev
RUN sed -i "s/\${VERSION}/${VERSION}/" open-api-3.yaml

# TARGETOS/TARGETARCH/TARGETVARIANT are provided by BuildKit. TARGETVARIANT is
# "v7" for linux/arm/v7 and empty otherwise, so stripping the leading "v" gives
# GOARM the value it expects.
# CGO_ENABLED=0 produces a static binary that runs on scratch.
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT#v} \
    go build -trimpath -ldflags="-s -w" -o main .

# --- Runtime base, shared by both targets (less than 10 MB) ---
FROM scratch AS runtime-base

# The binary needs the CA bundle to reach the upstream service over TLS.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/main /main

ENV PORT=8080
ENTRYPOINT ["/main"]


# --- Target 1: Standard runner (K8s, Docker, Raspberry Pi) ---
FROM runtime-base AS runner-standard
LABEL org.opencontainers.image.title="Abfallkalender API (Standard)"
LABEL org.opencontainers.image.description="Standard image for K8s, Docker or Raspberry Pi"
LABEL org.opencontainers.image.variant="standard"

EXPOSE 8080


# --- Target 2: AWS Lambda runner ---
# Without an explicit --platform, BuildKit resolves this image for the target
# platform automatically, so no per-architecture selection logic is needed.
FROM public.ecr.aws/awsguru/aws-lambda-adapter:1.1.0 AS adapter

FROM runtime-base AS runner-lambda
LABEL org.opencontainers.image.title="Abfallkalender API (Lambda)"
LABEL org.opencontainers.image.description="Optimized image for AWS Lambda execution"
LABEL org.opencontainers.image.variant="lambda"

# The AWS Lambda Web Adapter runs as a Lambda extension and forwards invocations
# to the HTTP server on $PORT.
COPY --from=adapter /lambda-adapter /opt/extensions/lambda-adapter

ENV RUST_LOG=info \
    AWS_LWA_ENABLE_COMPRESSION=true
