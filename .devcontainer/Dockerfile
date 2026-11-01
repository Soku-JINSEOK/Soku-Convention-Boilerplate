FROM node:24.22.2-bookworm@sha256:e7add2a490bb17d97b485b8641bd88455fff7839ade0f5eba493c3508124d413

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
