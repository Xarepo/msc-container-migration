package sftp

import (
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"

	chain_node "github.com/Xarepo/msc-container-migration/internal/chain/node"
	"github.com/Xarepo/msc-container-migration/internal/env"
	"github.com/Xarepo/msc-container-migration/internal/remote_target"
)

func isSymlink(fileName string) bool {
	fileInfo, err := os.Lstat(fileName)
	if err != nil {
		return false
	}

	return fileInfo.Mode()&os.ModeSymlink != 0
}

func CopyToRemote(node *chain_node.ChainNode, target *remote_target.RemoteTarget) {
	user := env.Getenv().SSH_USER
	password := env.Getenv().SSH_PASSWORD
	log.Debug().
		Str("User", user).
		Str("Password", password).
		Str("RemotePath", target.DumpPath).
		Str("Dump Name", node.Dump().Base()).
		Str("Target", target.Host).
		Msg("Copying to remote")

	clientConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: func(
			hostname string,
			remote net.Addr,
			key ssh.PublicKey,
		) error {
			return nil
		},
	}

	sshClient, err := ssh.Dial("tcp", target.FileTransferAddr(), clientConfig)
	if err != nil {
		log.Error().Str("Error", err.Error()).Msg("Failed to dial ssh")
	}

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		log.Error().Str("Error", err.Error()).Msg("Failed to create sftp client")
		return
	}
	defer sftpClient.Close()

	// Create dump directory on remote
	destDir := fmt.Sprintf("%s/%s", target.DumpPath, node.Dump().Base())
	log.Trace().Str("DestDir", destDir).Msg("Creating dump directory on remote")
	err = sftpClient.MkdirAll(destDir)
	if err != nil {
		log.Error().
			Str("Error", err.Error()).
			Str("DumpDir", destDir).
			Msg("Failed to create dump directory on remote")
		return
	}

	// Collect files
	files, err := filepath.Glob(fmt.Sprintf("%s/*", node.Dump().Path()))
	if err != nil {
		log.Error().Msg("Failed to collect files for transfer")
	}

	// Copy files to remote
	for _, file := range files {
		err := copyFile(&file, &destDir, node, sftpClient)
		if err != nil {
			log.Warn().
				Str("File", file).
				Str("Error", err.Error()).
				Msg("Failed to transfer file, chain is likely corrupt")
		}
	}
}

func copyFile(
	file, destDir *string,
	node *chain_node.ChainNode,
	sftpClient *sftp.Client,
) error {
	log.Trace().Str("File", *file).Msg("Transferring file")

	// Copy parent symlinks.
	// The only occurring symlinks in the dump directories should be the symlink
	// to the parent directory.
	if isSymlink(*file) {
		// TODO: This is nil at target host after migration, FIX
		oldName := node.GetPrev().Dump().ParentPath()
		log.Trace().
			Str("OldName", oldName).
			Str("NewName", *file).
			Msg("Creating parent symlink")
		err := sftpClient.Symlink(oldName, *file)
		if err != nil {
			return errors.Wrap(err, "Failed to create parent symlink")
		}
		return nil
	}

	dstFile, err := sftpClient.Create(path.Join(*destDir, filepath.Base(*file)))
	if err != nil {
		return errors.Wrap(err, "Failed to create remote file")
	}
	defer dstFile.Close()

	f, err := os.Open(*file)
	if err != nil {
		return errors.Wrap(err, "Failed to open file")
	}
	_, err = io.Copy(dstFile, f)
	if err != nil {
		return errors.Wrap(err, "Failed to write remote file")
	}

	return nil
}
