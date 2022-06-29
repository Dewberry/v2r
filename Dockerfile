FROM osgeo/gdal:alpine-normal-3.4.0 AS dev

COPY --from=golang:1.18.1-alpine /usr/local/go/ /usr/local/go/

ENV PATH="/usr/local/go/bin:${PATH}"

RUN apk add --no-cache git alpine-sdk

# Hot-Reloader
RUN go install github.com/githubnemo/CompileDaemon@v1.4.0

COPY ./ /app
WORKDIR /app

RUN go mod download
RUN go build main.go
ENTRYPOINT /root/go/bin/CompileDaemon --build="go build main.go" --command="./v2r"
