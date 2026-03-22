FROM golang:1.24-bookworm AS builder
WORKDIR /app
ADD . /app
RUN go build -o /dfloo-profile-go ./cmd/server
RUN go build -o /dfloo-f1-loader ./cmd/f1-loader

FROM debian:bookworm-slim

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

COPY --from=builder /dfloo-profile-go /dfloo-profile-go
COPY --from=builder /dfloo-f1-loader /dfloo-f1-loader
COPY --from=builder /app/internal/latex/templates /internal/latex/templates
COPY --from=builder /app/db/f1-data /db/f1-data

CMD ["/dfloo-profile-go"]