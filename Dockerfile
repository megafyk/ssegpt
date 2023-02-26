FROM golang:1.20-alpine as BUILDER
WORKDIR /go/src/app
COPY . .
RUN go mod vendor
RUN GOOS=linux go build -o main
FROM alpine
EXPOSE 8080
WORKDIR /app
COPY --from=BUILDER /go/src/app/main /app
CMD /app/main