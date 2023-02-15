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

COPY --from=build /go/bin/icloud-photo-cli /bin/icloud-photo-cli

ENTRYPOINT ["/bin/icloud-photo-cli"]
CMD ["download"]