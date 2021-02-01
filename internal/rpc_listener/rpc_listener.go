package rpc_listener

type RPCListener interface {
	Listen(func(buf []byte))
}
