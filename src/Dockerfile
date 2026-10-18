FROM --platform=linux/amd64 golang:1.27 AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download 2>/dev/null; go mod tidy
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server .

FROM --platform=linux/amd64 alpine:latest
WORKDIR /app
COPY --from=build /app/server .
EXPOSE 8080
CMD ["./server"]
