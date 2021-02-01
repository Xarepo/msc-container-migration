package cli

import (
	"flag"
	"os"

	"github.com/rs/zerolog/log"

	. "github.com/Xarepo/msc-container-migration/internal/cli_command"
	"github.com/Xarepo/msc-container-migration/internal/cli_commands"
	"github.com/Xarepo/msc-container-migration/internal/udp_listener"
)

func Parse() CliCommand {
	// Run command
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	runContainerId := runCmd.String(
		"container-id",
		"",
		"the id of the container")
	runBundlePath := runCmd.String(
		"bundle-path",
		"",
		"the path to the oci-bundle")

	// Migrate command
	migrateCmd := flag.NewFlagSet("migrate", flag.ExitOnError)
	migrateContainerId := migrateCmd.String(
		"container-id",
		"",
		"the id of the container")

	// Listen command
	listenCmd := flag.NewFlagSet("listen", flag.ExitOnError)

	if len(os.Args) < 2 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		runCmd.Parse(os.Args[2:])
	case "migrate":
		migrateCmd.Parse(os.Args[2:])
	case "listen":
		listenCmd.Parse(os.Args[2:])
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}

	if runCmd.Parsed() {
		if *runContainerId == "" || *runBundlePath == "" {
			log.Error().Msg("Missing value")
			runCmd.PrintDefaults()
			os.Exit(1)
		}

		return cli_commands.Run{
			BundlePath:  runBundlePath,
			ContainerId: runContainerId,
		}
	}

	if migrateCmd.Parsed() {
		if *migrateContainerId == "" {
			log.Error().Msg("Missing value")
			migrateCmd.PrintDefaults()
			os.Exit(1)
		}

		return cli_commands.Migrate{
			ContainerId: migrateContainerId,
		}
	}

	if listenCmd.Parsed() {
		return cli_commands.Listen{RPCListener: udp_listener.UDPListener{}}
	}

	return nil
}
