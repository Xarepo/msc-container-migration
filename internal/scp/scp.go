package scp

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/bramvdbogaerde/go-scp"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
)

func CopyToRemote(dumpName string) {
	user := os.Getenv("SCP_USER")
	password := os.Getenv("SCP_PASSWORD")
	remotePath := os.Getenv("SCP_REMOTE_PATH")
	target := os.Getenv("MIGRATION_TARGET")
	log.Trace().
		Str("User", user).
		Str("Password", password).
		Str("RemotePath", remotePath).
		Str("Dump Name", dumpName).
		Str("Target", target).
		Msg("Copying to remote")

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	client := scp.NewClient(target, config)
	err := client.Connect()
	if err != nil {
		log.Error().Str("Error", err.Error()).Msg("Failed to copy files")
		// fmt.Println("Couldn't establish a connection to the remote server ", err)
		return
	}
	defer client.Close()

	f, err := os.Open(dumpName)
	if err != nil {
		log.Error().Str("Error", err.Error()).Msg("Failed to open file")
	}
	defer f.Close()

	err = client.CopyFile(
		f,
		fmt.Sprintf("%s/%s", remotePath, filepath.Base(dumpName)),
		"0655")
	if err != nil {
		log.Error().Str("Error", err.Error()).Msg("Error while copying file ")
	}
}
