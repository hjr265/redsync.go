package redsync_test

import (
	"net"

	"github.com/hjr265/redsync.go/redsync"
)

func ExampleMutex() {
	m, err := redsync.NewMutex("FlyingSquirrels", []net.Addr{
		&net.TCPAddr{Port: 63790},
		&net.TCPAddr{Port: 63791},
		&net.TCPAddr{Port: 63792},
		&net.TCPAddr{Port: 63793},
	})
	if err != nil {
		panic(err)
	}

	err = m.Lock()
	if err != nil {
		panic(err)
	}
	defer m.Unlock()

	// Output:
}
