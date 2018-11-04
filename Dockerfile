FROM scratch
MAINTAINER Kris Budde <Kris.Budde@gmail.com>

COPY rabbitmq_exporter /

EXPOSE 9419

CMD ["/rabbitmq_exporter"]