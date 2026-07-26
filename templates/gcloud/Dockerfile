# Placeholder container for the GCP template.
# Downstream consumers should replace this file with a real application container
# build. This image intentionally keeps a deterministic entrypoint for CI build
# verification.

FROM alpine:3.21.7@sha256:48b0309ca019d89d40f670aa1bc06e426dc0931948452e8491e3d65087abc07d

WORKDIR /app
EXPOSE 8080

RUN apk add --no-cache busybox-extras \
    && printf 'ok\n' > /app/index.html \
    && cp /app/index.html /app/health

CMD ["httpd", "-f", "-p", "8080", "-h", "/app"]
