package sftp

import (
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/sftp"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"

	"github.com/Xarepo/msc-container-migration/internal/remote_target"
)

func isSymlink(fileName string) bool {
	fileInfo, err := os.Lstat(fileName)
	if err != nil {
		return false
	}

	return fileInfo.Mode()&os.ModeSymlink != 0
}

func CopyToRemote(dumpName string, target *remote_target.RemoteTarget) {
	user := os.Getenv("SCP_USER")
	password := os.Getenv("SCP_PASSWORD")
	log.Debug().
		Str("User", user).
		Str("Password", password).
		Str("RemotePath", target.DumpPath()).
		Str("Dump Name", dumpName).
		Str("Target", target.Host()).
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

	destDir := fmt.Sprintf("%s/%s", target.DumpPath(), filepath.Base(dumpName))
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
	files, err := filepath.Glob(fmt.Sprintf("%s/*", dumpName))
	if err != nil {
		log.Error().Msg("Failed to collect files for transfer")
	}

	for _, file := range files {
		log.Trace().Str("File", file).Msg("Transferring file")

		// Skip any symlinks.
		// The only occurring symlinks in the dump directories should be the symlink
		// to the parent directory, but this should not be necessary to transfer.
		if isSymlink(file) {
			log.Trace().Str("File", file).Msg("File is symlink, skipping...")
			continue
		}

		dstFile, err := sftpClient.Create(path.Join(destDir, filepath.Base(file)))
		if err != nil {
			log.Error().
				Str("File", path.Join(destDir, filepath.Base(file))).
				Str("Error", err.Error()).
				Msg("Failed to create remote file")
			return
		}
		defer dstFile.Close()

		f, err := os.Open(file)
		if err != nil {
			log.Error().Str("File", file).Msg("Failed to open file")
			continue
		}
		_, err = io.Copy(dstFile, f)
		if err != nil {
			log.Error().Str("File", file).Str("Error", err.Error()).Msg("Failed to write file")
			continue
		}
	}
}
