FROM golang:1.9

ADD . /go/src/github.com/AplaProject/go-apla

WORKDIR /go/src/github.com/AplaProject/go-apla

RUN CGO_ENABLED=1 GOOS=linux go build -o /srv/apla .

ENTRYPOINT /srv/apla