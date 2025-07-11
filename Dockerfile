FROM golang:1.23-bullseye AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o tubecast.o .

FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update -qq && \
    apt-get install -y --no-install-recommends \
        ca-certificates curl ffmpeg python3 python3-pip && \
    rm -rf /var/lib/apt/lists/*

RUN python3 -m pip install --no-cache-dir --break-system-packages \
        yt-dlp internetarchive          # latest PyPI versions

WORKDIR /app
COPY --from=builder /src/tubecast.o ./tubecast.o
COPY --from=builder /src .

RUN mkdir -p docs/feed

ENTRYPOINT ["./tubecast.o"]
CMD ["-sync"]