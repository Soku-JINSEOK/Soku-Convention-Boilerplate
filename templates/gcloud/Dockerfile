# Placeholder container for the GCP template.
# Downstream consumers should replace this file with a real application container
# build. This image intentionally keeps a deterministic entrypoint for CI build
# verification.

FROM alpine:3.24.3@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a

WORKDIR /app
EXPOSE 8080

RUN apk add --no-cache busybox-extras \
    && printf 'ok\n' > /app/index.html \
    && cp /app/index.html /app/health

CMD ["httpd", "-f", "-p", "8080", "-h", "/app"]
