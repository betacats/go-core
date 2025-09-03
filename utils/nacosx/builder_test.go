package nacosx

import (
	"fmt"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/rest"
)

type ZeroConfig struct {
	rest.RestConf
}

func TestNacos(t *testing.T) {
	nacosx := NewBuilder().
		WithServerAddr("localhost", 8848).
		WithNamespace("public").
		WithAuth("USER", "PASSWORD").
		WithDataID("vet").
		Execute()
	fmt.Println(nacosx.config.IPAddr)

	var c ZeroConfig
	nacosx.MustLoad(&c)

	fmt.Println("xxxxx", c.Name)

	nacosx.ListenConfig(func(namespace, group, dataID, data string) {
		nacosx.MustLoad(&c)
		fmt.Println("lister....c.Name...", c.Name)
	})

	time.Sleep(30 * time.Second)
}
