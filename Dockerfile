FROM debian:latest
RUN apt-get update
RUN apt-get install wget -y
RUN apt-get upgrade -y
RUN wget https://go.dev/dl/go1.21.7.linux-amd64.tar.gz
RUN tar -xvf go1.21.7.linux-amd64.tar.gz -C /usr/local
WORKDIR /app
COPY . .
RUN /usr/local/go/bin/go mod download
RUN /usr/local/go/bin/go mod verify
RUN /usr/local/go/bin/go build
CMD ["/app/github_report"]