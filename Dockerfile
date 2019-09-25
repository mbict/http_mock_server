FROM golang:latest as builder
MAINTAINER Michael Boke <michael@mbict.nl>
WORKDIR /app
ADD main.go main.go
ADD go.mod go.mod
RUN go mod download
RUN CGO_ENABLED=0 go build -a -ldflags '-s -w -extldflags "-static"' -o svc .

FROM scratch AS app
WORKDIR /
MAINTAINER Michael Boke <michael@mbict.nl>
COPY --from=builder /app/svc /svc
COPY --from=builder /etc/mime.types /etc/mime.types
EXPOSE 8080
CMD ["./svc"]