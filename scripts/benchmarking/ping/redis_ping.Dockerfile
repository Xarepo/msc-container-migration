FROM redis

RUN apt-get -y update && apt-get -y install iputils-ping
