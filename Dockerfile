FROM golang:alpine AS builder
WORKDIR /app
ADD . /app
RUN go build -o /dfloo-profile-go

FROM golang:alpine
COPY --from=builder /dfloo-profile-go /dfloo-profile-go
CMD ["/dfloo-profile-go"]