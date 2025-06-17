# syntax=docker/dockerfile:1

ARG REGISTRY="cgr.dev"

### Base
FROM --platform=$BUILDPLATFORM ${REGISTRY}/chainguard/wolfi-base:latest AS base

LABEL org.opencontainers.image.authors="GitGuardian SRE Team <support@gitguardian.com>"
LABEL org.opencontainers.image.title="GGBridge"
LABEL org.opencontainers.image.description="Connect your on-prem VCS with the GitGuardian Platform"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/GitGuardian/ggbridge"

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

ENV GGBRIDGE_SSL_CERT_DIR="/etc/ggbridge/ssl/certs"
ENV GGBRIDGE_SSL_CERT_FILE="${GGBRIDGE_SSL_CERT_DIR}/ca-bundle.crt"
ENV GGBRIDGE_SSL_PRIVATE_CERT_DIR="/etc/ggbridge/ssl/private"
ENV GGBRIDGE_PRIVATE_SSL_CERT_FILE="${GGBRIDGE_SSL_PRIVATE_CERT_DIR}/ca-bundle.crt"

RUN apk add --no-cache \
  bash \
  busybox \
  nginx \
  nginx-mod-stream

RUN mkdir -p $GGBRIDGE_SSL_CERT_DIR && \
  chown 65532:0 $GGBRIDGE_SSL_CERT_DIR && \
  chmod 775 $GGBRIDGE_SSL_CERT_DIR && \
  mkdir -p $GGBRIDGE_SSL_PRIVATE_CERT_DIR && \
  chown 65532:0 $GGBRIDGE_SSL_PRIVATE_CERT_DIR && \
  chmod 775 $GGBRIDGE_SSL_PRIVATE_CERT_DIR && \
  mkdir -p /opt/ggbridge && \
  chown 0:0 /opt/ggbridge && \
  chmod 775 /opt/ggbridge

COPY --chown=0:0 --chmod=0755 docker/scripts/run.sh /opt/ggbridge/run.sh
COPY --chown=0:0 --chmod=0644 docker/nginx/nginx.conf /etc/ggbridge/nginx.conf


### Builder
FROM base AS builder

RUN apk add --no-cache \
  bash \
  build-base \
  curl \
  git \
  go


### WSTunnel
FROM builder AS wstunnel

ARG WSTUNNEL_VERSION="10.4.3"
ENV WSTUNNEL_VERSION=$WSTUNNEL_VERSION
RUN curl -fsSL https://github.com/erebe/wstunnel/releases/download/v${WSTUNNEL_VERSION}/wstunnel_${WSTUNNEL_VERSION}_${TARGETOS}_${TARGETARCH}.tar.gz | \
  tar xvzf - -C /usr/bin wstunnel && \
  chmod 755 /usr/bin/wstunnel
USER 65532


### Build
FROM builder AS build

WORKDIR /build
COPY main.go go.mod go.sum .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
  go build -o ggbridge -ldflags "-w" .


### Dev
FROM builder AS dev

RUN apk add --no-cache \
  bind-tools \
  openssl \
  net-tools \
  vim

COPY --chown=0:0 --chmod=0444 docker/nginx/nginx.conf /etc/nginx/nginx.conf
COPY --link --from=wstunnel --chmod=755 /usr/bin/wstunnel /usr/bin/wstunnel


### Shell
FROM base AS shell

USER 0

RUN apk add --no-cache \
  bind-tools \
  curl \
  net-tools \
  openssl

COPY --link --from=wstunnel --chmod=755 /usr/bin/wstunnel /usr/bin/wstunnel
COPY --link --from=build --chmod=755 /build/ggbridge /usr/bin/ggbridge

USER 65532

STOPSIGNAL SIGQUIT

ENTRYPOINT ["/opt/ggbridge/run.sh"]
CMD ["client"]


### Prod
FROM base AS prod

COPY --link --from=wstunnel --chmod=755 /usr/bin/wstunnel /usr/bin/wstunnel
COPY --link --from=build --chmod=755 /build/ggbridge /usr/bin/ggbridge

ENTRYPOINT ["/opt/ggbridge/run.sh"]
CMD ["client"]
