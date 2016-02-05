package redsync_test

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/hjr265/redsync.go/redsync"
	"github.com/stvp/tempredis"
)

var servers []*tempredis.Server
var pools []redsync.Pool

func TestMain(m *testing.M) {
	for i := 0; i < 4; i++ {
		server, err := tempredis.Start(tempredis.Config{})
		if err != nil {
			panic(err)
		}
		defer server.Term()
		servers = append(servers, server)
	}
	pools = makeTestPools()
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
			m, err := redsync.NewMutexWithGenericPool("RedsyncMutex", pools)
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

func TestMutexWithRedSync(t *testing.T) {
	done := make(chan bool)
	chErr := make(chan error)

	rs := redsync.NewWithGenericPool(pools)
	for i := 0; i < 4; i++ {
		go func() {
			m := rs.NewMutex("RedsyncMutex2")
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

func makeTestPools() []redsync.Pool {
	pools := []redsync.Pool{}
	for _, server := range servers {
		func(server *tempredis.Server) {
			pools = append(pools, &redis.Pool{
				MaxIdle:     3,
				IdleTimeout: 240 * time.Second,
				Dial: func() (redis.Conn, error) {
					return redis.Dial("unix", server.Socket())
				},
				TestOnBorrow: func(c redis.Conn, t time.Time) error {
					_, err := c.Do("PING")
					return err
				},
			})
		}(server)
	}
	return pools
}
