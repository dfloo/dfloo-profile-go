FROM golang:alpine AS builder
WORKDIR /app
ADD . /app
RUN go build -o /dfloo-profile-go ./cmd/server

FROM golang:alpine

RUN apk add --no-cache \
    texlive \
    texlive-luatex \
    texlive-most \
    ca-certificates \
    && rm -rf /var/cache/apk/* \
    && rm -rf /usr/share/doc/* \
    && rm -rf /usr/share/man/* \
    && rm -rf /usr/share/texmf-dist/doc/* \
    && find /usr/share/texmf-dist -name "*.pdf" -delete \
    && find /usr/share/texmf-dist -name "*.dvi" -delete \
    && rm -rf /tmp/*

WORKDIR /app

COPY --from=builder /dfloo-profile-go /dfloo-profile-go
COPY --from=builder /app/internal/latex/templates /internal/latex/templates

CMD ["/dfloo-profile-go"]