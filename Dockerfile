FROM node:22-bookworm-slim AS ui

WORKDIR /src/ui
COPY ui/package.json ui/pnpm-lock.yaml ./
RUN corepack enable && corepack prepare pnpm@9.12.0 --activate && pnpm install --frozen-lockfile

COPY ui .
COPY pkg /src/pkg
RUN EMBED_UI=True pnpm build

FROM golang:1.26-bookworm AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=ui /src/pkg/template/vizb-ui.gen.go ./pkg/template/vizb-ui.gen.go

ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o /vizb .

FROM gcr.io/distroless/static-debian12:nonroot

# Release wiring may replace this build stage with the GoReleaser Linux binary.
COPY --from=build /vizb /vizb

ENTRYPOINT ["/vizb"]
CMD ["serve", "--host", "0.0.0.0", "--port", "8080"]
