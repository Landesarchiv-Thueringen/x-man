FROM docker.io/golang
WORKDIR /xman
RUN apt-get update
RUN apt-get install -y libxml2 libxml2-dev
COPY go.mod go.sum ./
RUN go mod download
COPY data/ca-certificates /usr/local/share/ca-certificates
RUN update-ca-certificates
ARG UID
ARG GID
RUN mkdir message_store && chown $UID:$GID message_store
COPY . ./
RUN go build cmd/server.go
USER $UID:$GID
CMD ["./server"]