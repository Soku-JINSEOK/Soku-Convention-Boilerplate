# Placeholder container for the GCP template.
# Downstream consumers should replace this file with a real application container
# build. This image intentionally keeps a deterministic entrypoint for CI build
# verification.

FROM alpine:3.24.1@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b

WORKDIR /app
EXPOSE 8080

RUN apk add --no-cache busybox-extras \
    && printf 'ok\n' > /app/index.html \
    && cp /app/index.html /app/health

CMD ["httpd", "-f", "-p", "8080", "-h", "/app"]
