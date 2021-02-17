FROM ubuntu:20.10

WORKDIR /app

# Add the binaries to the path
ENV PATH /app:$PATH

RUN apt-get update && apt-get install -y \
	git \
	golang \
	ca-certificates \
	runc \
	criu \
	ssh

COPY docker/docker-entrypoint.sh .
RUN chmod +x docker-entrypoint.sh

# Download go dependencies
COPY go.mod go.sum ./
RUN go mod download

# Allow root login and set password for root to root.
# This will also be the ssh password
RUN sed -ir 's/^#*PermitRootLogin .*/PermitRootLogin yes/' /etc/ssh/sshd_config
RUN echo "root:root" | chpasswd

COPY . .

RUN go build cmd/msc.go

ENTRYPOINT ["sh", "docker-entrypoint.sh"]
