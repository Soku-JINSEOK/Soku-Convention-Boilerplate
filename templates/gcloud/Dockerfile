# Placeholder container for the GCP template.
# Downstream consumers should replace this file with a real application container
# build. This image intentionally keeps a deterministic entrypoint for CI build
# verification.

FROM alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6

WORKDIR /app
EXPOSE 8080

RUN apk add --no-cache busybox-extras \
    && printf 'ok\n' > /app/index.html \
    && cp /app/index.html /app/health

CMD ["httpd", "-f", "-p", "8080", "-h", "/app"]
