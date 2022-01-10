FROM mongo:4.4

ENV MONGO_INITDB_ROOT_USERNAME admin
ENV MONGO_INITDB_ROOT_PASSWORD 04aba2b0e9a1fcd4

COPY logindb.sh /opt/
COPY dump.sh /opt/
COPY restore.sh /opt/
COPY init.js /docker-entrypoint-initdb.d/
RUN sed -i 's/\r//g' /opt/*.sh
RUN sed -i 's/\r//g' /docker-entrypoint-initdb.d/*
RUN chmod 777 /opt/*