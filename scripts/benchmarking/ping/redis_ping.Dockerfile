FROM redis:6.0.10

RUN apt-get -y update && apt-get -y install iputils-ping
