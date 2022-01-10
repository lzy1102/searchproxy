FROM alpine
#RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
#RUN set -x \
#    && apk add --no-cache tzdata \
#    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
#    && echo "Asia/Shanghai" > /etc/timezone
COPY start.sh /opt/
WORKDIR /opt/
RUN chmod +x *
#RUN rm -rf /opt/Dockerfile
CMD ["./start.sh"]
#CMD ["./task -scan=scanproxy -mod=true"]