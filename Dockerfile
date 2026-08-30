FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o url-shortener .

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/url-shortener .
EXPOSE 8080
CMD ["./url-shortener"]
