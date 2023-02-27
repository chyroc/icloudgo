# icloudgo

Access Apple iCloud via go, go port of pyicloud.

## Download iCloud Photos

### By Docker

```shell
docker run \
  -e ICLOUD_USERNAME=your_icloud_username \
  -e ICLOUD_PASSWORD=your_icloud_password \
  -e ICLOUD_COOKIE_DIR=/icloud_cookie \
  -e ICLOUD_DOMAIN=com \
  -e ICLOUD_OUTPUT=/icloud_photos \
  -e ICLOUD_ALBUM= \
  -e ICLOUD_THREAD_NUM=10 \
  -e ICLOUD_AUTO_DELETE=true \
  -e ICLOUD_STOP_FOUND_NUM=50 \
  -v /path/to/your/cookie:/icloud_cookie \
  -v /path/to/your/photos:/icloud_photos \
  ghcr.io/chyroc/icloud-photo-cli:0.9.0 download
```

### By Go

- **Install**

```shell
go install github.com/chyroc/icloudgo/icloud-photo-cli@latest
```

- **Usage**

```shell
NAME:
   icloud-photo-cli download

USAGE:
   icloud-photo-cli download [command options] [arguments...]

DESCRIPTION:
   download photos

OPTIONS:
   --username value, -u value                          apple id username [$ICLOUD_USERNAME]
   --password value, -p value                          apple id password [$ICLOUD_PASSWORD]
   --cookie-dir value, -c value                        cookie dir [$ICLOUD_COOKIE_DIR]
   --domain value, -d value                            icloud domain(com,cn) (default: com) [$ICLOUD_DOMAIN]
   --output value, -o value                            output dir (default: "./iCloudPhotos") [$ICLOUD_OUTPUT]
   --album value, -a value                             album name, if not set, download all albums [$ICLOUD_ALBUM]
   --stop-found-num stop-found-num, -s stop-found-num  stop download when found stop-found-num photos have been downloaded (default: 0) [$ICLOUD_STOP_FOUND_NUM]
   --thread-num value, -t value                        thread num, if not set, means 1 (default: 1) [$ICLOUD_THREAD_NUM]
   --auto-delete, --ad                                 auto delete photos after download (default: true) [$ICLOUD_AUTO_DELETE]
   --help, -h                                          show help
```


## Upload iCloud Photos

### By Docker

```shell
docker run \
  -i \
  -e ICLOUD_USERNAME=your_icloud_username \
  -e ICLOUD_PASSWORD=your_icloud_password \
  -e ICLOUD_COOKIE_DIR=/icloud_cookie \
  -e ICLOUD_DOMAIN=com \
  -e ICLOUD_FILE=/icloud_photos/filepath \
  -v /path/to/your/cookie:/icloud_cookie \
  -v /path/to/your/photos:/icloud_photos \
  ghcr.io/chyroc/icloud-photo-cli:0.9.0 upload
```

### By Go

- **Install**

```shell
go install github.com/chyroc/icloudgo/icloud-photo-cli@latest
```

- **Usage**

```shell
NAME:
   icloud-photo-cli upload

USAGE:
   icloud-photo-cli upload [command options] [arguments...]

DESCRIPTION:
   upload photos

OPTIONS:
   --username value, -u value    apple id username [$ICLOUD_USERNAME]
   --password value, -p value    apple id password [$ICLOUD_PASSWORD]
   --cookie-dir value, -c value  cookie dir [$ICLOUD_COOKIE_DIR]
   --domain value, -d value      icloud domain(com,cn) (default: com) [$ICLOUD_DOMAIN]
   --file value, -f value        file path [$ICLOUD_FILE]
   --help, -h                    show help
```
