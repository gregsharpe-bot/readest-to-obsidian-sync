FROM golang:1.22.12-bookworm AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -trimpath -ldflags='-s -w' -o /out/readest-obsidian-sync ./cmd/readest-obsidian-sync

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/readest-obsidian-sync /readest-obsidian-sync
USER nonroot:nonroot
ENTRYPOINT ["/readest-obsidian-sync"]
CMD ["run"]
