module github.com/johnwyles/vrddt-droplets

go 1.12

require (
	cloud.google.com/go v0.37.4
	github.com/golang/protobuf v1.3.1 // indirect
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.1
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/peter-jozsa/jsonpath v0.0.0-20180904092139-e43d4062dda8
	github.com/rs/cors v1.6.0
	github.com/sirupsen/logrus v1.4.1
	github.com/streadway/amqp v0.0.0-20190404075320-75d898a42a94
	github.com/stretchr/testify v1.3.0 // indirect
	go.opencensus.io v0.20.2 // indirect
	golang.org/x/net v0.0.0-20190419010253-1f3472d942ba // indirect
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a // indirect
	golang.org/x/sys v0.0.0-20190419153524-e8e3143a4f4a // indirect
	google.golang.org/api v0.3.2
	google.golang.org/appengine v1.5.0 // indirect
	google.golang.org/genproto v0.0.0-20190418145605-e7d98fc518a7 // indirect
	google.golang.org/grpc v1.20.1 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce
	gopkg.in/urfave/cli.v2 v2.0.0-20180128182452-d3ae77c26ac8
	gopkg.in/yaml.v2 v2.2.2 // indirect
)

replace gopkg.in/urfave/cli.v2 v2.0.0-20180128182452-d3ae77c26ac8 => github.com/johnwyles/cli v0.0.0-0.20190208003449-3a4ded93cc73
