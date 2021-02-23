FROM golang:1.14

COPY . /app

WORKDIR /app

RUN go build

ENTRYPOINT ["/app/jd"]
