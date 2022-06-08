FROM golang:1.18.1-bullseye


RUN apt-get update && \
    apt-get install -y build-essential 
  
RUN go install github.com/githubnemo/CompileDaemon@v1.4.0

COPY ./ /app
WORKDIR /app

RUN go mod download
