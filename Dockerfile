FROM golang:1.19 AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
		protobuf-compiler \
		nodejs \
		npm \
		&& rm -rf /var/lib/apt/lists/ \
		&& apt-get autoremove -y && apt-get autoclean -y

COPY . /src
WORKDIR /src

RUN npm --prefix ./tools/js ci
RUN make api-js
RUN GOPROXY=https://goproxy.cn make build

FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

COPY --from=builder /src/bin /app

WORKDIR /app

EXPOSE 8000
EXPOSE 9000
VOLUME /data/conf

CMD ["./server", "-conf", "/data/conf"]
