FROM node:22-alpine as builder

WORKDIR /app

COPY . .

RUN nm install && npm run build

FROM golang:1.23-alpine

WORKDIR /app

COPY --from=builder . .

RUN go mod tidy && go build -o ./app .

EXPOSE 80

CMD ./app
