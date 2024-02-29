FROM golang:alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd cmd
RUN CGO_ENABLED=0 GOOS=linux go build cmd/report.go

FROM alpine:edge
RUN echo https://dl-cdn.alpinelinux.org/alpine/edge/testing >> /etc/apk/repositories
RUN apk add --update typst font-noto
WORKDIR /app
COPY --from=build /app/report .
COPY template.typ .
# Preload Typst packages 
RUN typst compile template.typ || true
CMD ["./report"]