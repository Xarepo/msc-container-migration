package cli

import (
	"flag"
	"os"

	"github.com/rs/zerolog/log"

	. "github.com/Xarepo/msc-container-migration/internal/cli_command"
	"github.com/Xarepo/msc-container-migration/internal/cli_commands"
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

	// Join command
	joinCmd := flag.NewFlagSet("join", flag.ExitOnError)
	joinRemote := joinCmd.String(
		"remote",
		"",
		"the address of the remote source to join, in the form <host>:<rpcPort>",
	)

	if len(os.Args) < 2 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		runCmd.Parse(os.Args[2:])
	case "migrate":
		migrateCmd.Parse(os.Args[2:])
	case "join":
		joinCmd.Parse(os.Args[2:])
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

		return cli_commands.Migrate{ContainerId: migrateContainerId}
	}

	if joinCmd.Parsed() {
		if *joinRemote == "" {
			log.Error().Msg("Missing value")
			joinCmd.PrintDefaults()
			os.Exit(1)
		}
		return cli_commands.Join{Remote: *joinRemote}
	}

	return nil
}
