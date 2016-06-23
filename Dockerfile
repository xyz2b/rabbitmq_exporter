FROM scratch
MAINTAINER Kris Budde <Kris.Budde@gmail.com>

COPY rabbitmq_exporter /

EXPOSE      9090

CMD ["/rabbitmq_exporter"]