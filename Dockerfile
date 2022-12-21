FROM golang:1.18-buster AS build

WORKDIR /app
COPY . .

RUN go build -o /proxy

## 
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /proxy /proxy

EXPOSE 8080

ENTRYPOINT ["/proxy"]