FROM alpine

EXPOSE 8080

RUN apk --update upgrade && \
    apk add curl ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

COPY sc2a-service /root/sc2a-service
WORKDIR /root

ENTRYPOINT ["/root/sc2a-service"]