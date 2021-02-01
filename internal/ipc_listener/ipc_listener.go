package ipc_listener

type IPCListener interface {
	Listen(callback func(buf []byte))
}
