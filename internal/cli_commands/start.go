package cli_commands

import "fmt"

type Start struct {
	ContainerId *string
	BundlePath  *string
}

func (cmd Start) Execute() error {
	fmt.Println(*cmd.ContainerId)
	return nil
}
