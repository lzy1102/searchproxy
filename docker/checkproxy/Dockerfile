FROM alpine
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN set -x \
    && apk add --no-cache tzdata libpcap-dev masscan \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone
COPY . /opt/
WORKDIR /opt/
RUN chmod +x *
RUN rm -rf /opt/Dockerfile
CMD ["./checkproxy","-addr","http://restful-1:8080/api/get/list","-del","http://restful-1:8080/api/post/delete"]