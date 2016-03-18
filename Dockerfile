FROM scratch
MAINTAINER Kris Budde <Kris.Budde@gmail.com>

ADD rabbitmq_exporter_static /

EXPOSE 9090

CMD ["/rabbitmq_exporter_static"]



