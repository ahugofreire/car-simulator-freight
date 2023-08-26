FROM golang:1.19 as builder

RUN  apt update && apt install -y git ca-certificates

WORKDIR /app

COPY . .

CMD [ "tail", "-f", "/dev/null" ]