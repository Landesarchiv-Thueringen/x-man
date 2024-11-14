FROM docker.io/golang:1.23 AS build
RUN apt-get update && apt-get install -y libxml2-dev
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN go build cmd/server.go

FROM docker.io/debian:bookworm-slim
RUN apt-get update && \
    apt-get install -y --no-install-recommends libxml2 && \
    rm -rf /var/lib/apt/lists/*
WORKDIR /xman
COPY xsd ./xsd
COPY --from=build /build/server .
CMD ["./server"]