FROM golang:alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git tzdata

# Cache go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binaries
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker ./cmd/worker
RUN CGO_ENABLED=0 GOOS=linux go build -o /simulator ./cmd/simulator

# -------------------------
# API Image
# -------------------------
FROM gcr.io/distroless/static-debian12 AS api
WORKDIR /
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /api /api
EXPOSE 8080
EXPOSE 9090
USER nonroot:nonroot
ENTRYPOINT ["/api"]

# -------------------------
# Worker Image
# -------------------------
FROM gcr.io/distroless/static-debian12 AS worker
WORKDIR /
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /worker /worker
EXPOSE 9090
USER nonroot:nonroot
ENTRYPOINT ["/worker"]

# -------------------------
# Simulator Image
# -------------------------
FROM gcr.io/distroless/static-debian12 AS simulator
WORKDIR /
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /simulator /simulator
USER nonroot:nonroot
ENTRYPOINT ["/simulator"]
