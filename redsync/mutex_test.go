package redsync_test

import (
	"math/rand"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hjr265/redsync.go/redsync"
	"github.com/stvp/tempredis"
)

var addrs = []net.Addr{
	&net.TCPAddr{Port: 63790},
	&net.TCPAddr{Port: 63791},
	&net.TCPAddr{Port: 63792},
	&net.TCPAddr{Port: 63793},
}

var servers []*tempredis.Server

func TestMain(m *testing.M) {
	servers = make([]*tempredis.Server, len(addrs))
	for i, addr := range addrs {
		parts := strings.Split(addr.String(), ":")
		port := parts[1]
		server, err := tempredis.Start(tempredis.Config{"port": port})
		if err != nil {
			panic(err)
		}
		defer server.Term()
		servers[i] = server
	}
	result := m.Run()
	for _, server := range servers {
		server.Term()
	}
	os.Exit(result)
}

func TestMutex(t *testing.T) {
	done := make(chan bool)
	chErr := make(chan error)

	for i := 0; i < 4; i++ {
		go func() {
			m, err := redsync.NewMutex("RedsyncMutex", addrs)
			if err != nil {
				chErr <- err
				return
			}

			f := 0
			for j := 0; j < 32; j++ {
				err := m.Lock()
				if err == redsync.ErrFailed {
					f++
					if f > 2 {
						chErr <- err
						return
					}
					continue
				}
				if err != nil {
					chErr <- err
					return
				}

				time.Sleep(1 * time.Millisecond)

				m.Unlock()

				time.Sleep(time.Duration(rand.Int31n(128)) * time.Millisecond)
			}
			done <- true
		}()
	}
	for i := 0; i < 4; i++ {
		select {
		case <-done:
		case err := <-chErr:
			t.Fatal(err)
		}
	}
}
