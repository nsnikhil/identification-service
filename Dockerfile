FROM golang:alpine as builder
WORKDIR /identification-service
RUN apk update && apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o identification-service cmd/*.go

FROM scratch
COPY --from=builder /identification-service/identification-service .
ENTRYPOINT ["./identification-service"]