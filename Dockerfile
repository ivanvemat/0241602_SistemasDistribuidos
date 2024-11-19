FROM golang:1.22.6

RUN mkdir -p /home/logapp

COPY . /home/logapp

WORKDIR /home/logapp

RUN go mod download
RUN go install github.com/cloudflare/cfssl/cmd/cfssl@latest
RUN go install github.com/cloudflare/cfssl/cmd/cfssljson@latest

ENV CONFIG_PATH=/home/logapp/.logapp
RUN make init
RUN make gencert
RUN make copy_model
RUN make copy_policy

CMD [ "go", "test", "./...", "-v" ]