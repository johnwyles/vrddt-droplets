module github.com/johnwyles/vrddt-droplets

go 1.12

replace gopkg.in/urfave/cli.v2 v2.0.0-20180128182452-d3ae77c26ac8 => github.com/johnwyles/cli v0.0.0-0.20190208003449-3a4ded93cc73

require (
	cloud.google.com/go v0.37.0
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.0
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/peter-jozsa/jsonpath v0.0.0-20180904092139-e43d4062dda8
	github.com/sirupsen/logrus v1.4.0
	github.com/streadway/amqp v0.0.0-20190312002841-61ee40d2027b
	github.com/stretchr/testify v1.3.0 // indirect
	golang.org/x/crypto v0.0.0-20190325154230-a5d413f7728c // indirect
	golang.org/x/net v0.0.0-20190328230028-74de082e2cca // indirect
	golang.org/x/sys v0.0.0-20190329044733-9eb1bfa1ce65 // indirect
	google.golang.org/api v0.2.0
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce
	gopkg.in/urfave/cli.v2 v2.0.0-20180128182452-d3ae77c26ac8
)
