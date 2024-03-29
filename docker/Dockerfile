FROM ubuntu:20.10

WORKDIR /app

# RUN apk update && apk add git make
RUN apt-get update && apt-get install -y \
	git \
	make \
	golang \
	go-md2man \
	ca-certificates

# Compile oci-runtime-tools
ENV GOPATH=/go
RUN go get -v github.com/opencontainers/runtime-tools; exit 0 
RUN cd $GOPATH/src/github.com/opencontainers/runtime-tools && make && make install
RUN oci-runtime-tool generate --args "sh" --args "/count.sh" \
	--linux-namespace-remove network  > config.json

COPY . .

RUN go build cmd/msc.go

FROM ubuntu:20.10
WORKDIR /app

# Copy binaries
COPY --from=0 /app/msc .
COPY --from=0 /app/config.json .

RUN apt-get update && apt-get install -y runc criu ssh 

# Add the binaries to the path
ENV PATH /app:$PATH

COPY docker/docker-entrypoint.sh .
RUN chmod +x docker-entrypoint.sh

# Allow root login and set password for root to root.
# This will also be the ssh password
RUN sed -ir 's/^#*PermitRootLogin .*/PermitRootLogin yes/' /etc/ssh/sshd_config
RUN echo "root:root" | chpasswd

ENTRYPOINT ["sh", "docker-entrypoint.sh"]
