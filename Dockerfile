FROM golang:1.19 AS build

ENV GOPATH /go
WORKDIR /app

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/icloud-photo-cli ./icloud-photo-cli/main.go

RUN strip /go/bin/icloud-photo-cli
RUN test -e /go/bin/icloud-photo-cli

FROM alpine:latest

LABEL org.opencontainers.image.source=https://github.com/chyroc/icloudgo
LABEL org.opencontainers.image.description="Operate iCloud Photos."
LABEL org.opencontainers.image.licenses="Apache-2.0"

ENV ICLOUD_USERNAME=""
ENV ICLOUD_PASSWORD=""
ENV ICLOUD_COOKIE_DIR="/icloud_cookie"
ENV ICLOUD_DOMAIN="cn"
ENV ICLOUD_OUTPUT="/icloud_photos"
ENV ICLOUD_ALBUM=""
ENV ICLOUD_THREAD_NUM="10"
ENV ICLOUD_AUTO_DELETE="true"
ENV ICLOUD_STOP_FOUND_NUM="50"
ENV ICLOUD_FOLDER_STRUCTURE="2006/01/02"
ENV ICLOUD_FILE_STRUCTURE="id"
ENV ICLOUD_WITH_LIVE_PHOTO="true"

COPY --from=build /go/bin/icloud-photo-cli /bin/icloud-photo-cli

ENTRYPOINT ["/bin/icloud-photo-cli"]
CMD ["help"]