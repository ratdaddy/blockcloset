package config

import (
	"flag"
	"os"
	"strconv"
	"strings"
)

type envVal string

const (
	EnvProduction  envVal = "production"
	EnvStaging     envVal = "staging"
	EnvDevelopment envVal = "development"
	EnvTest        envVal = "test"
)

type logFormatVal string

const (
	LogJSON   logFormatVal = "json"
	LogPretty logFormatVal = "pretty"
)

type logVerbosityVal string

const (
	LogVerbose logVerbosityVal = "verbose"
	LogConcise logVerbosityVal = "concise"
)

var (
	AppEnv             envVal
	LogFormat          logFormatVal
	LogVerbosity       logVerbosityVal
	FlatbedPort        int
	GantryAddr         string
	PutObjectChunkSize int = 8192
)

func Init() {
	AppEnv = parseEnv(os.Getenv("APP_ENV"))

	switch AppEnv {
	case EnvDevelopment, EnvTest:
		LogFormat = LogPretty
		LogVerbosity = LogConcise
	default:
		LogFormat = LogJSON
		LogVerbosity = LogVerbose
	}

	FlatbedPort = 8080

	if v := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_FORMAT"))); v != "" {
		switch v {
		case "json":
			LogFormat = LogJSON
		case "pretty":
			LogFormat = LogPretty
		}
	}

	if v := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_VERBOSITY"))); v != "" {
		switch v {
		case "verbose":
			LogVerbosity = LogVerbose
		case "concise":
			LogVerbosity = LogConcise
		}
	}

	if v := strings.TrimSpace(os.Getenv("FLATBED_PORT")); v != "" {
		if port, err := strconv.Atoi(v); err == nil && port > 0 && port < 65536 {
			FlatbedPort = port
		}
	}

	if v := strings.TrimSpace(os.Getenv("FLATBED_GANTRY_ADDR")); v != "" {
		GantryAddr = v
	}

	if v := strings.TrimSpace(os.Getenv("PUT_OBJECT_CHUNK_SIZE")); v != "" {
		if size, err := strconv.Atoi(v); err == nil && size > 0 {
			PutObjectChunkSize = size
		}
	}
}

func parseEnv(v string) envVal {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "production":
		return EnvProduction
	case "staging":
		return EnvStaging
	case "test":
		return EnvTest
	case "development":
		return EnvDevelopment
	case "":
		if flag.Lookup("test.v") != nil {
			return EnvTest
		}
		return EnvDevelopment
	default:
		return EnvDevelopment
	}
}
