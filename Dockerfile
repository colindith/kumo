FROM golang:1.12.5

WORKDIR /usr/src/app
COPY . /usr/src/app/.

ENTRYPOINT ["/usr/src/app/entrypoint.sh"]

EXPOSE 8080