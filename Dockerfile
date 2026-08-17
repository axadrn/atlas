# Build natively on the builder's platform and cross compile for the target,
# so the arm64 image does not need slow QEMU emulation.
FROM --platform=$BUILDPLATFORM golang:1.25 AS build
ARG TARGETOS TARGETARCH
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Generate templ code with the exact version go.mod pins
RUN go install github.com/a-h/templ/cmd/templ@$(go list -m -f '{{.Version}}' github.com/a-h/templ)
RUN templ generate

# Tailwind standalone CLI, same major as the local toolchain
RUN ARCH=$(uname -m) && \
  if [ "$ARCH" = "x86_64" ]; then \
  TAILWIND_URL="https://github.com/tailwindlabs/tailwindcss/releases/download/v4.3.3/tailwindcss-linux-x64"; \
  elif [ "$ARCH" = "aarch64" ]; then \
  TAILWIND_URL="https://github.com/tailwindlabs/tailwindcss/releases/download/v4.3.3/tailwindcss-linux-arm64"; \
  else \
  echo "Unsupported architecture: $ARCH"; exit 1; \
  fi && \
  wget -q -O tailwindcss "$TAILWIND_URL" && \
  chmod +x tailwindcss
RUN ./tailwindcss -i ./assets/css/globals.css -o ./assets/css/output.css --minify

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w" -o main .

FROM alpine:3.20
WORKDIR /app

RUN apk add --no-cache ca-certificates

ENV GO_ENV=production
ENV PORT=8090
# Mount persistent storage and set DATABASE_PATH to a path inside that volume.
ENV DATABASE_PATH=/data/atlas.db
# NOINDEX is intentionally unset: the default is noindex on.
# Set NOINDEX=0 in the deploy env to allow search engines.

COPY --from=build /app/main .
COPY --from=build /app/assets ./assets

EXPOSE 8090

CMD ["./main"]
