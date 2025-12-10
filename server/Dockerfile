FROM docker.io/golang:1.25 AS build
ARG XMAN_VERSION=${XMAN_VERSION}
RUN apt-get update && apt-get install -y libxml2-dev
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN go build -o server -ldflags "-X main.version=${XMAN_VERSION}" ./cmd

FROM docker.io/debian:bookworm-slim
RUN apt-get update && \
    apt-get install -y --no-install-recommends libxml2 && \
    rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY xsd ./xsd
COPY --from=build /build/server .
CMD ["./server"]