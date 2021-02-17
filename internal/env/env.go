// package env provides a static environment, that should always contain valid
// values after calling the Init function.
package env

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/rs/zerolog/log"
)

type _env struct {
	LOG_LEVEL              string
	DUMP_PATH              string
	SCP_USER, SCP_PASSWORD string
	RPC_PORT               int
	CRIU_TCP_ESTABLISHED   bool
	DUMP_INTERVAL          int
	PING_INTERVAL          int
}

var env _env

func Init() error {
	var err error
	log.Trace().Msg("Initializing environment")
	env.LOG_LEVEL = getString("LOG_LEVEL", "info")
	env.DUMP_PATH = getString("DUMP_PATH", "/dumps")

	env.SCP_USER, err = getStringRequired("SCP_USER")
	if err != nil {
		return err
	}
	env.SCP_PASSWORD, err = getStringRequired("SCP_PASSWORD")
	if err != nil {
		return err
	}

	env.RPC_PORT, err = getInt("RPC_PORT", 1234)
	if err != nil {
		return err
	}

	env.DUMP_INTERVAL, err = getInt("DUMP_INTERVAL", 5)
	if err != nil {
		return err
	}

	env.PING_INTERVAL, err = getInt("PING_INTERVAL", 1)
	if err != nil {
		return err
	}

	env.CRIU_TCP_ESTABLISHED, err = getBool("CRIU_TCP_ESTABLISHED", false)
	if err != nil {
		return err
	}

	return nil
}

func Getenv() _env {
	return env
}

func getString(name, defaultValue string) string {
	val := os.Getenv(name)
	if val == "" {
		log.Warn().
			Str("Variable", name).
			Str("DefaultValue", defaultValue).
			Msg("Environment variable not set, using default value")
		return defaultValue
	}
	return val
}

func getStringRequired(name string) (string, error) {
	val := os.Getenv(name)
	if val == "" {
		return "", errors.New(
			fmt.Sprintf("Required environment variable %s not set", name),
		)
	}
	return val, nil
}

func getInt(name string, defaultValue int) (int, error) {
	val := os.Getenv(name)
	if val == "" {
		log.Warn().
			Str("Variable", name).
			Int("DefaultValue", defaultValue).
			Msg("Environment variable not set, using default value")
		return defaultValue, nil
	}

	valInt, err := strconv.Atoi(val)
	if err != nil {
		return math.MaxInt32, errors.New(
			fmt.Sprintf(
				"Failed to parse int from environment variable %s: %s",
				name,
				err.Error(),
			),
		)
	}

	return valInt, nil
}

func getBool(name string, defaultValue bool) (bool, error) {
	val := os.Getenv(name)
	if val == "" {
		log.Warn().
			Str("Variable", name).
			Bool("DefaultValue", defaultValue).
			Msg("Environment variable not set, using default value")
		return defaultValue, nil
	}

	valBool, err := strconv.ParseBool(val)
	if err != nil {
		return false, errors.New(
			fmt.Sprintf(
				"Failed to parse bool from environment variable %s: %s",
				name,
				err.Error(),
			),
		)
	}
	return valBool, nil
}
