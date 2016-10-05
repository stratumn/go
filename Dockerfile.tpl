FROM stratumn/gobase:latest

MAINTAINER Stephan Florquin <stephan@paymium.com>

COPY LICENSE /
COPY dist/linux-amd64/{{CMD}} /usr/local/bin/

CMD ["{{CMD}}"]
