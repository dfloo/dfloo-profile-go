FROM golang:1.24-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download
COPY . /app
RUN go build -o /out/dfloo-profile-go ./cmd/server
RUN go build -o /out/dfloo-f1-loader ./cmd/f1-loader

FROM debian:bookworm-slim AS runtime-base

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    fontconfig \
    lmodern \
    texlive-binaries \
    texlive-luatex \
    texlive-latex-base \
    texlive-latex-recommended \
    texlive-latex-extra \
    texlive-fonts-recommended \
    && rm -rf /var/lib/apt/lists/* \
    && rm -rf /usr/share/doc/* \
    && rm -rf /usr/share/man/* \
    && rm -rf /tmp/*

ENV HOME=/tmp
ENV TEXMFCACHE=/tmp/texmf-cache
ENV TEXMFVAR=/tmp/texmf-cache

WORKDIR /app

COPY --from=builder /app/internal/latex/templates /internal/latex/templates

FROM runtime-base AS runtime-loader

COPY --from=builder /out/dfloo-f1-loader /dfloo-f1-loader
COPY --from=builder /app/db/f1-data /db/f1-data

CMD ["/dfloo-f1-loader", "--data-dir", "/db/f1-data"]

FROM runtime-base AS runtime-api

COPY --from=builder /out/dfloo-profile-go /dfloo-profile-go

CMD ["/dfloo-profile-go"]