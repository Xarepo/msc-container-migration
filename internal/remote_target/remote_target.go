package remote_target

import "fmt"

type RemoteTarget struct {
	host             string
	rpcPort          int
	dumpPath         string
	fileTransferPort int
}

func New(host string, rpcPort int, dumpPath string, fileTransferPort int) RemoteTarget {
	return RemoteTarget{
		host:             host,
		rpcPort:          rpcPort,
		dumpPath:         dumpPath,
		fileTransferPort: fileTransferPort,
	}
}

func (target RemoteTarget) RPCAddr() string {
	return fmt.Sprintf("%s:%d", target.host, target.rpcPort)
}

func (target RemoteTarget) FileTransferAddr() string {
	return fmt.Sprintf("%s:%d", target.host, target.fileTransferPort)
}

func (target RemoteTarget) DumpPath() string {
	return target.dumpPath
}

func (target RemoteTarget) Host() string {
	return target.host
}

func (target RemoteTarget) RPCPort() int {
	return target.rpcPort
}

func (target RemoteTarget) FileTransferPort() int {
	return target.fileTransferPort
}
