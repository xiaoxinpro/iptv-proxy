FROM --platform=${TARGETPLATFORM} golang:alpine AS builder

ENV CGO_ENABLED 0

WORKDIR /app

ADD go.mod .
RUN go mod download
COPY . .
RUN go build -ldflags "-w -s" -trimpath -o iptv-proxy ./src

FROM --platform=${TARGETPLATFORM} alpine

LABEL authors="Chishin"

ENV TZ=Asia/Shanghai
ENV WEB_HOST=""
ENV WEB_PORT=3006
ENV IPTV_SOURCE=""
ENV IPTV_TARGET=iptv.m3u
ENV IPTV_UA=""
ENV IPTV_REFRESH=3600

WORKDIR /app

COPY --from=builder /app/iptv-proxy .

RUN apk update --no-cache && apk add --no-cache ca-certificates tzdata && \
    chmod +x ./iptv-proxy

EXPOSE $WEB_PORT

CMD ["./iptv-proxy"]