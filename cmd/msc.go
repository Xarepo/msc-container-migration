package main

import (
	"context"
	"fmt"

	runc "github.com/containerd/go-runc"
)

func main() {
	cfg := &runc.Runc{}

	v, err := cfg.Version(context.Background())
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(v)
}
