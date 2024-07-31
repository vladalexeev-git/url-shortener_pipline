FROM golang:1.22.5 as builder

#RUN mkdir /build
#ADD . /build/
WORKDIR /build
COPY . .

RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache

# Сборка приложения с использованием кеша
RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache \
    go mod tidy

RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache \
    GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o url-shortener ./cmd/url-shortener/main.go


#FROM alpine:latest
#FROM debian:sid-slim
FROM ubuntu:22.04

WORKDIR /app

COPY --from=builder /build/config/* /app/config/
COPY --from=builder /build/url-shortener .

EXPOSE 8081

RUN chmod +x /app/url-shortener

CMD ["./url-shortener"]