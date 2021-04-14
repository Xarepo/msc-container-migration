// package env provides a static environment, that should always contain valid
// values after calling the Init function.
package env

import (
	"math"
	"os"
	"strconv"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type _env struct {
	ENABLE_CONTINOUS_DUMPING                         bool
	LOG_LEVEL                                        string
	DUMP_PATH                                        string
	SSH_USER, SSH_PASSWORD                           string
	RPC_PORT                                         int
	CRIU_TCP_ESTABLISHED                             bool
	DUMP_INTERVAL                                    int
	PING_INTERVAL, PING_TIMEOUT, PING_TIMEOUT_SOURCE int
	CHAIN_LENGTH                                     int
}

var env _env

// Default values for the optional environment variables.
const (
	_DEFAULT_ENABLE_CONTINOUS_DUMPING = true
	_DEFAULT_LOG_LEVEL                = "info"
	_DEFAULT_DUMP_PATH                = "/dumps"
	_DEFAULT_RPC_PORT                 = 1234
	_DEFAULT_DUMP_INTERVAL            = 5
	_DEFAULT_PING_INTERVAL            = 1
	_DEFAULT_PING_TIMEOUT             = 5
	_DEFAULT_CRIU_TCP_ESTABLISHED     = false
	_DEFAULT_PING_TIMEOUT_SOURCE      = 3
	_DEFAULT_CHAIN_LENGTH             = 3
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

	env.SSH_USER, err = getStringRequired("SSH_USER")
	if err != nil {
		return err
	}
	env.SSH_PASSWORD, err = getStringRequired("SSH_PASSWORD")
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

	env.CHAIN_LENGTH, err = getInt("CHAIN_LENGTH", _DEFAULT_CHAIN_LENGTH)
	if err != nil {
		return err
	}

	env.ENABLE_CONTINOUS_DUMPING, err = getBool(
		"ENABLE_CONTINOUS_DUMPING",
		_DEFAULT_ENABLE_CONTINOUS_DUMPING,
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
		return "", errors.Errorf("Required environment variable %s not set", name)
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
		return math.MaxInt32, errors.Wrapf(
			err,
			"Failed to parse int from environment variable %s",
			name,
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
		return false, errors.Wrapf(
			err,
			"Failed to parse bool from environment variable %s",
			name,
		)
	}
	return valBool, nil
}
