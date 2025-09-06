package config

import (
	"flag"
	"os"
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
	AppEnv           envVal
	LogFormat        logFormatVal
	LogVerbosity     logVerbosityVal
	EnableReflection bool
)

func Init() {
	AppEnv = parseEnv(os.Getenv("APP_ENV"))

	switch AppEnv {
	case EnvDevelopment, EnvTest:
		LogFormat = LogPretty
		LogVerbosity = LogConcise
		EnableReflection = true
	default:
		LogFormat = LogJSON
		LogVerbosity = LogVerbose
		EnableReflection = false
	}

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

	if v := strings.ToLower(strings.TrimSpace(os.Getenv("ENABLE_REFLECTION"))); v != "" {
		switch v {
		case "true", "True", "1", "yes", "on":
			EnableReflection = true
		case "false", "False", "0", "no", "off":
			EnableReflection = false
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
