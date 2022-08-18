FROM ubuntu:bionic
ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update && \
    apt-get dist-upgrade -y && \
    apt-get install -y --no-install-recommends \
      bison build-essential ca-certificates curl clang cmake flex llvm pkg-config tar && \
    rm -rf /var/cache/apt/* && \
    rm -rf /var/lib/apt/lists/* && \
    rm -rf /tmp/* && \
    rm -rf /var/tmp/*

ENV GO_VERSION 1.17.6
ENV GO_URL https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
ENV GO_SHA256 231654bbf2dab3d86c1619ce799e77b03d96f9b50770297c8f4dff8836fc8ca2
RUN curl -fsSL "$GO_URL" -o go.tar.gz && \
    echo "$GO_SHA256  go.tar.gz" | sha256sum -c - && \
    tar -C /usr/local -xzf go.tar.gz && \
    ln -s /usr/local/go/bin/go /usr/local/bin && \
    rm go.tar.gz

RUN mkdir -p /thrift
COPY build-thrift.sh /thrift/build.sh
