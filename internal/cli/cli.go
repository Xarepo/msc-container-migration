package cli

import (
	"flag"
	"os"

	. "github.com/Xarepo/msc-container-migration/internal/cli_command"
	"github.com/Xarepo/msc-container-migration/internal/cli_commands"
	"github.com/rs/zerolog/log"
)

func Parse() CliCommand {
	startCmd := flag.NewFlagSet("start", flag.ExitOnError)

	startContainerId := startCmd.String("container-id", "", "the id of the container")
	startBundlePath := startCmd.String("bundle-path", "", "the path to the oci-bundle")

	if len(os.Args) < 2 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "start":
		startCmd.Parse(os.Args[2:])
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}

	if startCmd.Parsed() {
		if *startContainerId == "" || *startBundlePath == "" {
			log.Error().Msg("Missing value")
			startCmd.PrintDefaults()
			os.Exit(1)
		}

		return cli_commands.Start{BundlePath: startBundlePath, ContainerId: startContainerId}
	}
	return nil
}
