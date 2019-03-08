module github.com/johnwyles/vrddt-droplets

go 1.12

replace gopkg.in/urfave/cli.v2 v2.0.0-20180128182452-d3ae77c26ac8 => github.com/johnwyles/cli v1.16.1-0.20190208003449-3a4ded93cc73

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.0
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/peter-jozsa/jsonpath v0.0.0-20180904092139-e43d4062dda8
	github.com/rs/zerolog v1.12.0 // indirect
	github.com/sirupsen/logrus v1.3.0
	github.com/spf13/cobra v0.0.3
	github.com/spf13/viper v1.3.1
	github.com/streadway/amqp v0.0.0-20190225234609-30f8ed68076e
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce
	gopkg.in/urfave/cli.v2 v2.0.0-20180128182452-d3ae77c26ac8
)
