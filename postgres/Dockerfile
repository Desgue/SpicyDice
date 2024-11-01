FROM postgres:16 

ENV POSTGRES_DB postgres
ENV POSTGRES_USER postgres
ENV POSTGRES_PASSWORD p4ssw0rd

COPY ./migrations/ /docker-entrypoint-initdb.d/ 