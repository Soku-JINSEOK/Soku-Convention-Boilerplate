FROM node:24.18.0-bookworm@sha256:5711a0d445a1af54af9589066c646df387d1831a608226f4cd694fc59e745059

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
