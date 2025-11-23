# Does not work for arm
#   # Use multi stage build to# minimize generated docker images size
#   # see: https://docs.docker.com/develop/develop-images/multistage-build/
#
#   # Step 1: create multi stage assets builder
#   # HINT: up to now parcel does not support arm -> https://github.com/parcel-bundler/parcel/issues/5812
#   # .github actions configuration is using multi arch target plattform
#   # assets builder is only used to create assets, using the official node:alpine image
#   FROM node:alpine AS assets
#
#   # Create app directory
#   WORKDIR /app
#
#   # Install python and other dependencies via apk
#   RUN apk update && apk add python3 g++ make && rm -rf /var/cache/apk/*
#
#   # Install app dependencies
#   # A wildcard is used to ensure both package.json AND package-lock.json are copied where available (npm@5+)
#   COPY package*.json /app/
#   COPY src/frontend /app/src/frontend
#
#   # Create package-lock.json
#   #RUN npm i --package-lock-only
#
#   # Make a clean npm install and only install modules needed for production
#   RUN npm ci
#
#   # Build assets
#   RUN npm run build

# --- Pre-Stage: Fetch Adapters for different architectures ---
FROM --platform=linux/arm64 public.ecr.aws/awsguru/aws-lambda-adapter:0.9.1 AS adapter-source-arm64
FROM --platform=linux/amd64 public.ecr.aws/awsguru/aws-lambda-adapter:0.9.1 AS adapter-source-amd64

# --- Step 2: create multi stage backend builder (about 800 MB) ---
FROM golang:1.25 AS builder
LABEL stage=intermediate
RUN go version

WORKDIR /app

COPY main.go /app/
COPY go.* /app/
COPY open-api-3.yaml /app
COPY src/backend /app/src/backend
#COPY --from=assets /app/dist /app/dist

RUN go mod download
RUN go mod tidy # prevent missing go.sum entry for module

# CGO_ENABLED=0   -> Disable interoperate with C libraries -> speed up build time!
# -a              -> force rebuilding of packages that are already up-to-date.
# -o app          -> force to build an executable app file

ARG BUILDPLATFORM
ARG TARGETPLATFORM

RUN echo "I am running on $BUILDPLATFORM, building for $TARGETPLATFORM"

# set version in open-api-3.yaml (provided via build-arg)
ARG VERSION
RUN sed -i "s/\${VERSION}/${VERSION}/" open-api-3.yaml

# Build Logic
RUN if [ "$TARGETPLATFORM" = "linux/arm/v7" ] ; then \
        echo "I am building linux/arm/v7 with CGO_ENABLED=0 GOARCH=arm GOARM=7" ; \
        env CGO_ENABLED=0 GOARCH=arm GOARM=7 go build -a -o main . ; \
        echo "Build done" ; \
    fi

RUN if [ "$TARGETPLATFORM" = "linux/arm64" ] ; then \
        echo "I am building linux/arm64 with CGO_ENABLED=0 GOARCH=arm64 GOARM=7" ; \
        env CGO_ENABLED=0 GOARCH=arm64 GOARM=7 go build -a -o main . ; \
        echo "Build done" ; \
    fi

RUN if [ "$TARGETPLATFORM" = "linux/amd64" ] ; then \
        echo "I am building linux/amd64 with CGO_ENABLED=0 GOARCH=amd64" ; \
        env CGO_ENABLED=0 GOARCH=amd64 go build -a -o main . ; \
        echo "Build done" ; \
    fi

# --- Adapter Selection Logic in Builder Stage ---
# We prepare the correct adapter here, so we can simply COPY it in the final stage
COPY --from=adapter-source-arm64 /lambda-adapter /tmp/lambda-adapter-arm64
COPY --from=adapter-source-amd64 /lambda-adapter /tmp/lambda-adapter-amd64

RUN touch /app/lambda-adapter
RUN if [ "$TARGETPLATFORM" = "linux/arm64" ]; then \
      echo "Selecting ARM64 adapter"; \
      cp /tmp/lambda-adapter-arm64 /app/lambda-adapter; \
    elif [ "$TARGETPLATFORM" = "linux/amd64" ]; then \
      echo "Selecting AMD64 adapter"; \
      cp /tmp/lambda-adapter-amd64 /app/lambda-adapter; \
    else \
      echo "No adapter for $TARGETARCH (creating dummy)"; \
    fi
RUN chmod +x /app/lambda-adapter && ls -lh /app/lambda-adapter


# --- Step 3: Base Runtime (Common for all targets) ---
# create minimal executable image (less than 10 MB)
FROM scratch AS runtime-base
WORKDIR /root/

# copy the ca-certificate.crt from the build stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy app binary and config
COPY --from=builder /app/main .
COPY --from=builder /app/open-api-3.yaml .


# --- Target 1: Standard Runner (K8s, Docker, Raspi) ---
# This target DOES NOT include the adapter
FROM runtime-base AS runner-standard
LABEL org.opencontainers.image.title="Abfallkalender API (Standard)"
LABEL org.opencontainers.image.description="Standard image for K8s, Docker or Raspberry Pi"
LABEL org.opencontainers.image.variant="standard"

EXPOSE 8080
ENV PORT=8080
ENTRYPOINT ["./main"]


# --- Target 2: AWS Lambda Runner ---
# This target INCLUDES the adapter
FROM runtime-base AS runner-lambda
LABEL org.opencontainers.image.title="Abfallkalender API (Lambda)"
LABEL org.opencontainers.image.description="Optimized image for AWS Lambda execution"
LABEL org.opencontainers.image.variant="lambda"

# Add AWS Lambda Web Adapter as an extension
COPY --from=builder /app/lambda-adapter /opt/extensions/lambda-adapter

# Environment variables for AWS Lambda Web Adapter
# PORT must match the port your app listens on
ENV PORT=8080 \
    RUST_LOG=info \
    AWS_LWA_ENABLE_COMPRESSION=true

ENTRYPOINT ["./main"]