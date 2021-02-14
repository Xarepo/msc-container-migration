package remote_target

import "fmt"

type RemoteTarget struct {
	Host             string
	RPCPort          int
	DumpPath         string
	FileTransferPort int
}

func New(host string, rpcPort int, dumpPath string, fileTransferPort int) RemoteTarget {
	return RemoteTarget{
		Host:             host,
		RPCPort:          rpcPort,
		DumpPath:         dumpPath,
		FileTransferPort: fileTransferPort,
	}
}

func (target RemoteTarget) RPCAddr() string {
	return fmt.Sprintf("%s:%d", target.Host, target.RPCPort)
}

func (target RemoteTarget) FileTransferAddr() string {
	return fmt.Sprintf("%s:%d", target.Host, target.FileTransferPort)
}

// func (target RemoteTarget) DumpPath() string {
// 	return target.dumpPath
// }

// func (target RemoteTarget) Host() string {
// 	return target.host
// }

// func (target RemoteTarget) RPCPort() int {
// 	return target.rpcPort
// }

// func (target RemoteTarget) FileTransferPort() int {
// 	return target.fileTransferPort
// }
