FROM golang:1.13.8

COPY . /application

WORKDIR /application/cmd/back

RUN go build
RUN mv back /bin/back

COPY configs/back-config.yaml /etc/application/config.yaml
COPY api/swagger-ui /opt/application/swagger-ui
COPY assets/migrations /opt/application/assets/migrations

EXPOSE 80

CMD ["/bin/back", "-c", "/etc/application/config.yaml"]