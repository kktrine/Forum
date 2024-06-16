FROM golang:latest

ENV GOPATH=/

COPY ./ ./

# install psql
RUN apt-get update
RUN apt-get -y install postgresql-client
RUN chmod +x postgres.sh

# build service
RUN go get -d -v ./...
RUN go mod download
RUN go build -o forum ./cmd/server/main.go
EXPOSE 8080
CMD ["./commands"]