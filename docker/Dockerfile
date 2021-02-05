FROM golang:1.15.8-alpine3.12

WORKDIR /app

RUN apk update && apk add git make go-md2man

# RUN apt-get update && apt-get install -y \
# 	git \
# 	make \
# 	golang \
# 	go-md2man \
# 	ca-certificates

# # Compile oci-runtime-tools
ENV GOPATH=/go
RUN go get -v github.com/opencontainers/runtime-tools; exit 0 
RUN cd $GOPATH/src/github.com/opencontainers/runtime-tools && make && make install
RUN oci-runtime-tool generate --args "sh" --args "/count.sh" --linux-namespace-remove network > config.json

COPY . .

RUN go build cmd/msc.go

FROM alpine:3.12.3

WORKDIR /app

# Copy binaries
COPY --from=0 /app/msc .
COPY --from=0 /app/config.json .


# RUN apk update && apk add runc curl bzip2 make gcc git autoconf g++ libtool
RUN apk update && apk add runc curl bzip2 make libc6-dev-i386 gcc-multilib git autoconf g++ libtool


# RUN git clone https://github.com/google/protobuf.git protobuf && \
#  cd protobuf && \
#  ./autogen.sh && \
#  ./configure --prefix=`pwd`/../`uname -m`-linux-gnu && \
#  make && \
#  make install && \
#  cd ../..


RUN curl http://download.openvz.org/criu/criu-3.15.tar.bz2 -o curl.tar.bz2 && \
	bzip2 -d curl.tar.bz2 && \ 
	tar xvf curl.tar
RUN apk add \
	protobuf protobuf-dev \
	protobuf-c protobuf-c-dev \
	protoc \
	libprotobuf \
	linux-headers \
	libnet libnet-dev \
	libnl3 libnl3-dev \
	libcap libcap-dev \
	py3-ipaddress \
	asciidoc \
	xmlto \
	nftables nftables-dev \
	libbsd libbsd-dev \
	pkgconfig \
	iproute2
RUN cd criu-3.15 && make && make install

# Add the binaries to the path
ENV PATH /app:$PATH
RUN echo LOG_LEVEL=trace >> .env
RUN cat .env
CMD ["msc", "run", "--container-id", "counter", "--bundle-path", "."]

# FROM ubuntu:20.10
# WORKDIR /app

# # Copy binaries
# COPY --from=0 /app/msc .
# COPY --from=0 /app/config.json .

# RUN apt-get update && apt-get install -y runc criu

# # Add the binaries to the path
# ENV PATH /app:$PATH

# CMD ["msc", "run", "--container-id", "counter", "--bundle-path", "."]
