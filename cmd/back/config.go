package main

import (
	"io/ioutil"

	"github.com/go-pg/pg/v9"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server     Server            `yaml:"server"`
	Swagger    Swagger           `yaml:"swagger"`
	DB         *pg.Options       `yaml:"db_config"`
	Migration  MigrationConfig   `yaml:"migration"`
	Centrifuge *CentrifugeConfig `yaml:"centrifuge"`
	Auth AuthConfig `yaml:"auth"`
}

type Swagger struct {
	Path string `yaml:"path"`
	Url  string `yaml:"url"`
}

type Server struct {
	GrpcAddress    string `yaml:"grpc_address"`
	GatewayAddress string `yaml:"gateway_address"`
	BasePath       string `yaml:"base_path"`
}

type AuthConfig struct {
	PrivateKey string `yaml:"private_key"`
	PublicKey string `yaml:"public_key"`
}

func Configure(fileName string) (Config, error) {
	var cnf Config
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return Config{}, err
	}
	err = yaml.Unmarshal(data, &cnf)
	if err != nil {
		return Config{}, err
	}

	return cnf, nil
}
