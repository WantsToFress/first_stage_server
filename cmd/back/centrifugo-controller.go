package main

import (
	"github.com/centrifugal/centrifuge-go"
	"github.com/dgrijalva/jwt-go"
)

type CentrifugeConfig struct {
	Addr       string            `yaml:"addr"`
	HMACSecret string            `yaml:"hmac_secret"`
	ClientId   string            `yaml:"client_id"`
	Centrifuge centrifuge.Config `yaml:"centrifuge"`
}

func NewCentrifugoClient(config *CentrifugeConfig) (*centrifuge.Client, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = jwt.MapClaims{
		"sub": config.ClientId,
	}
	tokenRaw, err := token.SignedString([]byte(config.HMACSecret))
	if err != nil {
		return nil, err
	}

	cent := centrifuge.New(config.Addr, centrifuge.DefaultConfig())
	cent.SetToken(tokenRaw)
	err = cent.Connect()
	if err != nil {
		return nil, err
	}
	return cent, nil
}
