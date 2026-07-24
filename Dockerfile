# syntax=docker/dockerfile:1

# ---------- Stage 1: build SPA ----------
# Mirror the repo layout so Vite's outDir (../../backend/web/dist) resolves correctly.
FROM oven/bun:1.3-alpine AS frontend
WORKDIR /repo/frontend/app

COPY frontend/app/package.json frontend/app/bun.lockb* ./
RUN --mount=type=cache,target=/root/.bun/install/cache \
    bun install --frozen-lockfile

COPY frontend/app/ ./
RUN mkdir -p /repo/backend/web/dist && bun run build


# ---------- Stage 2: build static Go binary ----------
FROM golang:1.26-alpine AS backend
WORKDIR /src

COPY backend/go.mod backend/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY backend/ ./
RUN rm -rf web/dist && mkdir -p web/dist
COPY --from=frontend /repo/backend/web/dist/ web/dist/

# Fully static, stripped, reproducible binary.
ARG TARGETOS TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w -buildid=" -o /out/server .

# Create the data directory here so it can be copied into the shell-less runtime
# image with the right ownership. A named volume mounted at /data inherits this
# ownership, so the nonroot user can write the SQLite file.
RUN mkdir -p /data


# ---------- Stage 3: runtime ----------
# distroless: no shell, no pkg manager, ~2MB. :nonroot runs as uid 65532.
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

WORKDIR /app
COPY --from=backend --chown=nonroot:nonroot /out/server /app/server
# Empty, nonroot-owned data dir; a mounted volume inherits this ownership.
COPY --from=backend --chown=nonroot:nonroot /data /data

USER nonroot:nonroot
EXPOSE 3000
ENV ADDR=":3000"
# SQLite database location. Mount a volume at /data to persist it.
ENV DB_PATH="/data/work_items.db"
VOLUME ["/data"]

# The binary probes itself in --healthcheck mode, so the shell-less distroless
# image needs no extra tooling.
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD ["/app/server", "--healthcheck"]

ENTRYPOINT ["/app/server"]
