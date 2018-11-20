FROM golang:latest

ARG PROTOC_VERSION=3.5.1

RUN apt-get update && \
      apt-get install unzip && \
      curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh && \
      curl -fo protoc.zip -sSL "https://github.com/google/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip" && \
      mkdir protoc && \
      unzip -q protoc.zip -d protoc && \
      cp ./protoc/bin/protoc /usr/local/bin/ && \
      cp -rf ./protoc/include/google /usr/local/include/ && \
      rm protoc.zip protoc -rf && \
      go get -u github.com/golang/protobuf/protoc-gen-go && \
      rm -rf ${GOPATH}/src/github.com/golang ${GOPATH}/pkg && \
      rm -rf /var/lib/apt/lists/*

WORKDIR /go/src/github.com/nametake/protoc-gen-gohttp
