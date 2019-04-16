#
# Build Stage
#
FROM golang:1.11-alpine as builder

ARG COMMAND_PATH=""
ARG COMMAND_SUFFIX=""
ARG VERSION_FLAG=""

ENV COMMAND_PATH=$COMMAND_PATH
ENV COMMAND_SUFFIX=$COMMAND_SUFFIX
ENV VERSION_FLAG=$VERSION_FLAG

WORKDIR $GOPATH/src/github.com/johnwyles/vrddt-droplets

COPY . ./

# -a -installsuffix cgo  -ldflags -v -ldflags "-extldflags \"-static\""
RUN CGO_ENABLED=0 GOOS=linux go build --ldflags -v --ldflags "$VERSION_FLAG" -o /vrddt-$COMMAND_SUFFIX $COMMAND_PATH


#
# Final Stage
#
FROM jrottenberg/ffmpeg:4.1-alpine

ARG COMMAND_PATH=""
ARG COMMAND_SUFFIX=""
ARG VERSION_FLAG=""

ENV COMMAND_PATH=$COMMAND_PATH
ENV COMMAND_SUFFIX=$COMMAND_SUFFIX
ENV DOCKERIZE_VERSION v0.6.1
ENV VERSION_FLAG=$VERSION_FLAG

COPY ./config/gcs/vrddt-225521.json /vrddt-225521.json

RUN apk add --no-cache openssl
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz
RUN mkdir -p /app

COPY --from=builder /vrddt-$COMMAND_SUFFIX /app/vrddt-$COMMAND_SUFFIX
RUN chmod +x /app/vrddt-$COMMAND_SUFFIX

ENTRYPOINT /app/vrddt-$COMMAND_SUFFIX