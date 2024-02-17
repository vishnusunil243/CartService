#Build stage
FROM golang:1.21-bullseye AS builder

RUN apt-get update

WORKDIR /cart

COPY . .
 
COPY .env .env

RUN go mod download

RUN go build -o ./out/dist ./cmd

#production stage

FROM busybox

RUN mkdir -p /cart/out/dist

COPY --from=builder /cart/out/dist /cart/out/dist

COPY --from=builder /cart/.env /cart/out

WORKDIR /cart/out/dist

EXPOSE 8083

CMD ["./dist"]