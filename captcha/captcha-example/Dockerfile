# syntax=docker/dockerfile:1

FROM node:24-bookworm-slim AS frontend

WORKDIR /workspace

COPY captcha/captcha-example/frontend/package.json captcha/captcha-example/frontend/package.json
COPY captcha/captcha-example/frontend/yarn.lock captcha/captcha-example/frontend/yarn.lock

RUN cd captcha/captcha-example/frontend && npm install --no-package-lock

COPY captcha/captcha-example/frontend captcha/captcha-example/frontend
COPY captcha/captcha-example/internal/captcha/html captcha/captcha-example/internal/captcha/html

RUN cd captcha/captcha-example/frontend && npm run build


FROM golang:1.24-bookworm AS builder

WORKDIR /workspace

COPY pkg/sdk pkg/sdk
COPY captcha/captcha-example captcha/captcha-example
COPY --from=frontend /workspace/captcha/captcha-example/internal/captcha/html captcha/captcha-example/internal/captcha/html

WORKDIR /workspace/captcha/captcha-example

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/captcha-example .


FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /out/captcha-example /app/captcha-example

ENTRYPOINT ["/app/captcha-example"]
