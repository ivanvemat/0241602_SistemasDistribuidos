FROM golang:1.22.6

RUN mkdir -p /home/logapp

COPY . /home/logapp

WORKDIR /home/logapp

RUN go mod download

ENV CONFIG_PATH=/home/logapp/.logapp
RUN make init

CMD [ "go", "test", "./...", "-v" ]