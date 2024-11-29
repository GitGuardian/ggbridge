# syntax=docker/dockerfile:1

ARG REGISTRY="cgr.dev"

### Base
FROM --platform=$BUILDPLATFORM ${REGISTRY}/chainguard/wolfi-base:latest AS base

LABEL org.opencontainers.image.authors="GitGuardian SRE Team <support@gitguardian.com>"

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

RUN apk add --no-cache \
  curl


### WSTunnel
FROM base AS wstunnel

ARG WSTUNNEL_VERSION="10.1.6"
ENV WSTUNNEL_VERSION=$WSTUNNEL_VERSION
RUN curl -fsSL https://github.com/erebe/wstunnel/releases/download/v${WSTUNNEL_VERSION}/wstunnel_${WSTUNNEL_VERSION}_${TARGETOS}_${TARGETARCH}.tar.gz | \
  tar xvzf - -C /usr/bin wstunnel && \
  chmod 755 /usr/bin/wstunnel
USER 65532


### Builder
FROM base AS builder

RUN apk add --no-cache \
  bash \
  git \
  go


### Build
FROM builder AS build

WORKDIR /build
COPY main.go go.mod go.sum .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
  go build -o ggbridge -ldflags "-w" .


### Dev
FROM builder AS dev

RUN apk add --no-cache \
  nano \
  nginx \
  nginx-mod-stream \
  openssl \
  vim

COPY --chown=0:0 --chmod=0444 docker/files/nginx/nginx.conf /etc/nginx/nginx.conf
COPY --link --from=wstunnel --chmod=755 /usr/bin/wstunnel /usr/bin/wstunnel


### NGINX
FROM ${REGISTRY}/chainguard/nginx:latest-dev AS nginx

LABEL org.opencontainers.image.authors="GitGuardian SRE Team <support@gitguardian.com>"
LABEL org.opencontainers.image.description="Connect your on-prem VCS with the GitGuardian Platform"

USER 0

RUN apk add --no-cache \
  nginx-mod-stream

USER 65532


### Debug
FROM nginx AS debug

USER 0

RUN apk add --no-cache \
  bash \
  curl \
  openssl

COPY --chown=0:0 --chmod=0444 docker/files/nginx/nginx.conf /etc/nginx/nginx.conf
COPY --link --from=wstunnel --chmod=755 /usr/bin/wstunnel /usr/bin/wstunnel
COPY --link --from=build --chmod=755 /build/ggbridge /usr/bin/ggbridge


USER 65532

ENTRYPOINT ["/usr/bin/ggbridge"]
CMD ["client"]


### Prod
FROM ${REGISTRY}/chainguard/nginx:latest AS prod

LABEL org.opencontainers.image.authors="GitGuardian SRE Team <support@gitguardian.com>"
LABEL org.opencontainers.image.description="Connect your on-prem VCS with the GitGuardian Platform"

COPY --chown=0:0 --chmod=0444 docker/files/nginx/nginx.conf /etc/nginx/nginx.conf
COPY --link --from=nginx --chmod=755 /usr/lib/nginx/modules /usr/lib/nginx/modules
COPY --link --from=nginx --chmod=755 /etc/nginx/modules /etc/nginx/modules
COPY --link --from=nginx --chmod=755 /etc/nginx/stream.d /etc/nginx/stream.d
COPY --link --from=wstunnel --chmod=755 /usr/bin/wstunnel /usr/bin/wstunnel
COPY --link --from=build --chmod=755 /build/ggbridge /usr/bin/ggbridge

ENTRYPOINT ["/usr/bin/ggbridge"]
CMD ["client"]
