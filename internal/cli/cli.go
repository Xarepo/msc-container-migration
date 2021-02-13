package cli

import (
	"github.com/Xarepo/msc-container-migration/internal/cli/cli_commands"

	"github.com/alecthomas/kong"
)

var cli struct {
	Run     cli_commands.Run     `kong:"cmd,help:'Run a container'"`
	Join    cli_commands.Join    `kong:"cmd,help:'Join a cluster'"`
	Migrate cli_commands.Migrate `kong:"cmd,help:'Migrate a container'"`
}

type CliCommand interface {
	Execute() error
}

func Parse() CliCommand {
	ctx := kong.Parse(&cli, kong.UsageOnError())
	switch ctx.Command() {
	case "run <container-id>":
		return cli.Run
	case "join <remote>":
		return cli.Join
	case "migrate <container-id>":
		return cli.Migrate
	default:
		panic(ctx.Command())
	}
}
