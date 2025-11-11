# =========================
# Stage 1: Build the binary
# =========================
FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum Makefile ./
RUN make dep

COPY . .

RUN make build

# =========================
# STAGE 2: runtime
# =========================
FROM alpine:3.22

RUN addgroup -g 1000 -S appgroup && adduser -u 1000 -S appuser -G appgroup

RUN mkdir /app && chown appuser:appgroup /app
WORKDIR /app

COPY --from=builder /app/go-template /app/go-template

USER appuser

ENTRYPOINT ["/app/go-template"]
