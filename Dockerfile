FROM ghostbaby/alpine:3.8
MAINTAINER  Herman Zhu <zhuhuijun@gmail.com>

RUN mkdir /app
WORKDIR /app
COPY cfs-broker /app/cfs-broker

EXPOSE      8080
ENTRYPOINT  [ "/app/cfs-broker" ]
