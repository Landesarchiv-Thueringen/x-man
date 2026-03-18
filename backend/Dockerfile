FROM docker.io/golang:alpine3.23 AS build
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN go build -o xman ./cmd

FROM docker.io/golang:alpine3.23
WORKDIR /app
COPY --from=build /build/xman .
CMD ["./xman"]