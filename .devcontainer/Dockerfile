FROM node:24.17.0-bookworm@sha256:733e1c06ada118ed9f6133a31aa1290be6929664026fb28821500437c61f2c6f

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
