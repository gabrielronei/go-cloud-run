FROM golang:1.23
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server .
EXPOSE 8080
CMD ["./server"]
