package rpc_listener

type RPCListener interface {
	Listen(port int, callback func(buf []byte, remoteAddr string))
}
