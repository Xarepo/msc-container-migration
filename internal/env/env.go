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
	LOG_LEVEL                                        string
	DUMP_PATH                                        string
	SCP_USER, SCP_PASSWORD                           string
	RPC_PORT                                         int
	CRIU_TCP_ESTABLISHED                             bool
	DUMP_INTERVAL                                    int
	PING_INTERVAL, PING_TIMEOUT, PING_TIMEOUT_SOURCE int
}

var env _env

// Default values for the optional environment variables.
const (
	_DEFAULT_LOG_LEVEL            = "info"
	_DEFAULT_DUMP_PATH            = "/dumps"
	_DEFAULT_RPC_PORT             = 1234
	_DEFAULT_DUMP_INTERVAL        = 5
	_DEFAULT_PING_INTERVAL        = 1
	_DEFAULT_PING_TIMEOUT         = 5
	_DEFAULT_CRIU_TCP_ESTABLISHED = false
	_DEFAULT_PING_TIMEOUT_SOURCE  = 3
)

// Initialize the environment.
// If any required variables are missing, then an error will be returned and
// the environment should not be used as the rest of the environment will be
// uninitialized.
func Init() error {
	var err error
	log.Trace().Msg("Initializing environment")
	env.LOG_LEVEL = getString("LOG_LEVEL", _DEFAULT_LOG_LEVEL)
	env.DUMP_PATH = getString("DUMP_PATH", _DEFAULT_DUMP_PATH)

	env.SCP_USER, err = getStringRequired("SCP_USER")
	if err != nil {
		return err
	}
	env.SCP_PASSWORD, err = getStringRequired("SCP_PASSWORD")
	if err != nil {
		return err
	}

	env.RPC_PORT, err = getInt("RPC_PORT", _DEFAULT_RPC_PORT)
	if err != nil {
		return err
	}

	env.DUMP_INTERVAL, err = getInt("DUMP_INTERVAL", _DEFAULT_DUMP_INTERVAL)
	if err != nil {
		return err
	}

	env.PING_INTERVAL, err = getInt("PING_INTERVAL", _DEFAULT_PING_INTERVAL)
	if err != nil {
		return err
	}
	env.PING_TIMEOUT, err = getInt("PING_TIMEOUT", _DEFAULT_PING_TIMEOUT)
	if err != nil {
		return err
	}
	if env.PING_TIMEOUT < env.PING_INTERVAL {
		log.Warn().
			Int("PING_TIMEOUT", env.PING_TIMEOUT).
			Int("PING_INTERVAL", env.PING_INTERVAL).
			Msg("PING_TIMEOUT is less than PING_INTERVAL." +
				" Any node joining the cluster will always restore")
	}
	env.PING_TIMEOUT_SOURCE, err = getInt(
		"PING_TIMEOUT_SOURCE",
		_DEFAULT_PING_TIMEOUT_SOURCE)
	if err != nil {
		return err
	}

	env.CRIU_TCP_ESTABLISHED, err = getBool(
		"CRIU_TCP_ESTABLISHED",
		_DEFAULT_CRIU_TCP_ESTABLISHED,
	)
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
