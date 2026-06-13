# syntax=docker/dockerfile:1

# --- Build stage ---------------------------------------------------------
FROM golang:1.24 AS build

WORKDIR /src

# Cache module downloads.
COPY go.mod go.sum ./
RUN go mod download

# Build the static binary.
COPY . .
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# --- Runtime stage -------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /
COPY --from=build /out/server /server

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/server"]
