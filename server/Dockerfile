FROM docker.io/golang:1.23
WORKDIR /xman
RUN apt-get update && apt-get install -y libxml2 libxml2-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN go build cmd/server.go
CMD ["./server"]