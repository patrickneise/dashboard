FROM mcr.microsoft.com/devcontainers/go:2-1.25-bookworm

ARG AIR_VERSION=1.63.0
ARG GOLANGCI_LINT_VERSION=1.61.0
ARG GOOSE_VERSION=3.24.0
ARG MIGRATE_VERSION=4.19.0
ARG SQLC_VERSION=1.27.0
ARG TEMPL_VERSION=0.3.960
ARG TARGETARCH=amd64

USER root

# Install sqlite3 (for DB debugging)
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
        sqlite3 \
    && rm -rf /var/lib/apt/lists/*

# Install golangci-lint
RUN curl -sSfL "https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh" \
      | sh -s -- -b /usr/local/bin "v${GOLANGCI_LINT_VERSION}"

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
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@v${SQLC_VERSION} && \
    go install -tags 'sqlite' github.com/golang-migrate/migrate/v4/cmd/migrate@v${MIGRATE_VERSION} && \
    go install github.com/a-h/templ/cmd/templ@v${TEMPL_VERSION} && \
    go install github.com/pressly/goose/v3/cmd/goose@v${GOOSE_VERSION} && \
    go install github.com/air-verse/air@v${AIR_VERSION}
