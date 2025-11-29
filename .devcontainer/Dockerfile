FROM mcr.microsoft.com/devcontainers/go:2-1.25-bookworm

ARG MIGRATE_VERSION=4.19.0
ARG TARGETARCH=amd64

USER root

# Install sqlite3 (for DB debugging)
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
    sqlite3

# Install Tailwind CSS standalone CLI
# Use TARGETARCH (provided by Docker) to get the correct binary (x64 or arm64)
RUN case ${TARGETARCH} in \
        amd64) ARCH_SUFFIX="x64" ;; \
        arm64) ARCH_SUFFIX="arm64" ;; \
        *) echo "Unsupported architecture: ${TARGETARCH}"; exit 1 ;; \
    esac \
    && curl -sL "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-${ARCH_SUFFIX}" -o /usr/local/bin/tailwindcss \
    && chmod +x /usr/local/bin/tailwindcss

USER vscode

# Install Go-based CLIs
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest && \
    go install -tags 'sqlite' github.com/golang-migrate/migrate/v4/cmd/migrate@latest && \
    go install github.com/a-h/templ/cmd/templ@latest && \
    go install github.com/pressly/goose/v3/cmd/goose@latest && \
    go install github.com/air-verse/air@latest