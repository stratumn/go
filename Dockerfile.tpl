FROM stratumn/gobase:v0.4.0-rc2

MAINTAINER Stephan Florquin <stephan@stratumn.com>

RUN addgroup -S -g 999 stratumn
RUN adduser -H -D -u 999 -G stratumn stratumn

COPY LICENSE /usr/local/stratumn/
COPY {{CMD}} /usr/local/stratumn/bin/

RUN mkdir /usr/local/bin
RUN ln -s /usr/local/stratumn/bin/{{CMD}} /usr/local/bin/{{CMD}}

USER stratumn

CMD ["{{CMD}}"]
