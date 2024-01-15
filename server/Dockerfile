FROM golang
WORKDIR /xman
RUN apt-get update
RUN apt-get install -y libxml2 libxml2-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
CMD ["go", "run", "cmd/server.go"]