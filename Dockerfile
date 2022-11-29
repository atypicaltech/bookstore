## Build
FROM golang:1.19-buster AS build

WORKDIR /app

# Install dependencies for Webloop (https://github.com/sourcegraph/webloop)
RUN apt-get update -y \
    && apt-get install --no-install-recommends -yq \
    software-properties-common \
    wget \
    build-essential \
    ca-certificates


COPY go.mod ./
COPY go.sum ./
COPY ./* ./
RUN go mod download \
    && go mod tidy \
    && go build -o /bookstore


## Deploy
FROM debian:buster

WORKDIR /

COPY --from=build /bookstore /bookstore

USER nonroot:nonroot
ENTRYPOINT ["/bookstore"]
