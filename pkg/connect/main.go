package connect

import "sync"

type ConnectClient interface {
	Status() (connected, intercept bool, err error)
	IsConnected() (*string, int, error)
	Connect(ns string) error
	Disconnect() error
	Intercept(name string, ns string) error
	Leave(name string, ns string) error
}

func NewConnect() ConnectClient {
	return NewKubeVpn()
}

var (
	client    ConnectClient
	singleTon = sync.Once{}
)

func SClient() ConnectClient {
	singleTon.Do(func() {
		client = NewKubeVpn()
	})

	return client
}
