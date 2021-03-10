FROM alpine
RUN apk --update upgrade && \
    apk add curl ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

ADD finneg.com.key /
ADD STAR_finneg_com.crt /

ADD application /
ADD server.yml /
WORKDIR /
CMD ["/application"]
EXPOSE 5000
