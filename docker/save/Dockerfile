FROM alpine
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN set -x \
    && apk add --no-cache tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone
COPY . /opt/
WORKDIR /opt/
RUN chmod 777 *
RUN rm -rf /opt/Dockerfile
CMD ./task -scan=save -mod=true