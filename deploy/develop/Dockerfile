FROM alpine:latest

ARG TARGETPLATFORM

EXPOSE 8080 6380 6060

LABEL maintainer="Dmitry Ponomarev <demdxx@gmail.com>"
LABEL service.name=redify

ENV SERVER_HTTP_LISTEN=:8080
ENV SERVER_REDIS_LISTEN=:6380
ENV SERVER_PROFILE_MODE=net
ENV SERVER_PROFILE_LISTEN=6060

COPY deploy/develop/example.config.yml /example.config.yml
COPY .build/${TARGETPLATFORM}/redify /redify

ENTRYPOINT [ "/redify", "--conf", "/example.config.yml" ]
