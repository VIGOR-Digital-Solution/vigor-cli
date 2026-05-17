# syntax=docker/dockerfile:1.7
# Distroless static — ~15 MB final image. Used by teammates who want vigor in CI
# without installing Go or Homebrew.

FROM gcr.io/distroless/static-debian12:nonroot
COPY vigor /usr/local/bin/vigor
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/vigor"]
