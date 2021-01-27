package cli_command

type CliCommand interface {
	Execute() error
}
