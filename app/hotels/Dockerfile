# ---------- Build stage ----------
FROM golang:1.25.5 AS builder

WORKDIR /src
COPY . .

RUN GOPROXY=https://goproxy.cn make build

# ---------- Runtime stage ----------
FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
        netbase \
        gettext-base \
        && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /src/bin/hotels /app/server

COPY ./configs /data/conf

COPY app/hotels/hotels_dataset.json /app/hotels_dataset.json

RUN chmod +x /app/server

EXPOSE 8201
EXPOSE 9201

VOLUME /data/conf

CMD ["./server", "-conf", "/data/conf"]
