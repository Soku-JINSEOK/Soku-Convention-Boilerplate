FROM node:24.20.0-bookworm@sha256:be23f54a88d34e8824c741b19b91064094f92c1c97b194144bfc8b50d67258e2

RUN apt-get update \
    && apt-get install --yes --no-install-recommends \
      bash \
      ca-certificates \
      docker.io \
      git \
      golang-go \
      maven \
      python3 \
      python3-pip \
      python3-venv \
      shellcheck \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace
