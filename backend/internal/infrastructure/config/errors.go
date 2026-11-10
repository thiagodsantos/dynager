package config

import "errors"

var (
	errConfigPathRequired             = errors.New("config path is required")
	errInvalidConfigPath              = errors.New("invalid config path")
	errCreateConfigDirectory          = errors.New("could not create config directory")
	errReadConnectionConfiguration    = errors.New("could not read connection configuration")
	errInvalidConnectionConfiguration = errors.New("connection configuration is invalid")
	errEncodeConnectionConfiguration  = errors.New("could not encode connection configuration")
	errWriteConnectionConfiguration   = errors.New("could not write connection configuration")
	errSecureConnectionConfiguration  = errors.New("could not secure connection configuration")
	errReplaceConnectionConfiguration = errors.New("could not replace connection configuration")
	errAddressMustBeLoopback          = errors.New("address must be loopback")
)
