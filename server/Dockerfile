FROM golang
WORKDIR /xman
RUN apt-get update
RUN apt-get install -y libxml2 libxml2-dev
COPY go.mod go.sum ./
RUN go mod download
COPY data/ca-certificates /usr/local/share/ca-certificates
RUN update-ca-certificates
COPY . ./
RUN go build cmd/server.go
CMD ["./server"]