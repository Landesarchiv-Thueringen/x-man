FROM docker.io/golang:alpine3.21 AS build
ARG XMAN_VERSION=${XMAN_VERSION}
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o report -ldflags "-X main.version=${XMAN_VERSION}" ./cmd

FROM docker.io/alpine:3.21
RUN apk add --no-cache typst font-noto
WORKDIR /app
COPY --from=build /build/report .
COPY *.typ .
# Preload Typst packages 
RUN typst compile submission-report.typ || true
CMD ["./report"]