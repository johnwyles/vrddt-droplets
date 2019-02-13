FROM golang:1.11 as builder
RUN mkdir /vrddt-droplets-src
WORKDIR /vrddt-droplets-src
COPY ./ .
RUN CGO_ENABLED=0 make setup all

FROM alpine:latest
RUN mkdir /app
WORKDIR /app
COPY --from=builder /vrddt-droplets-src/bin/vrddt-droplets ./
COPY --from=builder /vrddt-droplets-src/web ./web
EXPOSE 8080
CMD ["./vrddt-droplets"]
