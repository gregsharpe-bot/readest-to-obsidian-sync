FROM golang:1.22.12-bookworm AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -trimpath -ldflags='-s -w' -o /out/readest-obsidian-sync ./cmd/readest-obsidian-sync

FROM node:22-bookworm-slim

ARG OBSIDIAN_HEADLESS_VERSION=0.0.14
RUN npm install --global --omit=dev "obsidian-headless@${OBSIDIAN_HEADLESS_VERSION}" && npm cache clean --force
COPY --from=builder --chown=node:node /out/readest-obsidian-sync /usr/local/bin/readest-obsidian-sync
USER node
ENTRYPOINT ["readest-obsidian-sync"]
CMD ["run"]
