FROM node:24.21.0-bookworm@sha256:64af3819f9275802414d7cdc38c27e9d82bd564dec4d4da87d008255d36c63b4

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
