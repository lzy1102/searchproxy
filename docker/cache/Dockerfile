FROM redis:alpine
COPY redis.conf /usr/local/etc/redis/redis.conf
COPY login.sh /opt/
RUN chmod 777 /opt/*.sh
CMD ["redis-server", "/usr/local/etc/redis/redis.conf" ]