// Package redsync provides a Redis-based distributed mutual exclusion lock implementation as described in the blog post http://antirez.com/news/77.
//
// Values containing the types defined in this package should not be copied.
package redsync

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/fzzy/radix/redis"
)

const (
	DefaultExpiry = 8 * time.Second
	DefaultTries  = 16
	DefaultDelay  = 512 * time.Millisecond
	DefaultFactor = 0.01
)

var (
	ErrFailed = errors.New("failed to acquire lock")
)

type Locker interface {
	Lock() error
	Unlock()
}

// A Mutex is a mutual exclusion lock.
//
// Fields of a Mutex must not be changed after first use.
type Mutex struct {
	Name   string        // Resouce name
	Expiry time.Duration // Duration for which the lock is valid, DefaultExpiry if 0

	Tries int           // Number of attempts to acquire lock before admitting failure, DefaultTries if 0
	Delay time.Duration // Delay between two attempts to acquire lock, DefaultDelay if 0

	Factor float64 // Drift factor, DefaultFactor if 0

	Quorum int // Quorum for the lock, set to len(addrs)/2+1 by NewMutex()

	value string
	until time.Time

	nodes []*redis.Client
	nodem sync.Mutex
}

var _ = Locker(&Mutex{})

// NewMutex returns a new Mutex on a named resource connected to the Redis instances at given addresses.
func NewMutex(name string, addrs []net.Addr) (*Mutex, error) {
	if len(addrs) == 0 {
		panic("redsync: addrs is empty")
	}

	nodes := []*redis.Client{}
	for _, addr := range addrs {
		node, _ := redis.Dial(addr.Network(), addr.String())
		nodes = append(nodes, node)
	}

	return &Mutex{
		Name:   name,
		Quorum: len(addrs)/2 + 1,
		nodes:  nodes,
	}, nil
}

// Lock locks m.
// In case it returns an error on failure, you may retry to acquire the lock by calling this method again.
func (m *Mutex) Lock() error {
	m.nodem.Lock()
	defer m.nodem.Unlock()

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return err
	}
	value := base64.StdEncoding.EncodeToString(b)

	expiry := m.Expiry
	if expiry == 0 {
		expiry = DefaultExpiry
	}

	retries := m.Tries
	if retries == 0 {
		retries = DefaultTries
	}

	for i := 0; i < retries; i++ {
		n := 0
		start := time.Now()
		for _, node := range m.nodes {
			if node == nil {
				continue
			}

			reply := node.Cmd("set", m.Name, value, "nx", "px", int(expiry/time.Millisecond))
			if reply.Err != nil {
				continue
			}
			if reply.String() != "OK" {
				continue
			}
			n += 1
		}

		factor := m.Factor
		if factor == 0 {
			factor = DefaultFactor
		}

		until := time.Now().Add(m.Expiry - time.Now().Sub(start) - time.Duration(int64(float64(m.Expiry)*factor)) + 2*time.Millisecond)
		if n >= m.Quorum && time.Now().Before(until) {
			m.value = value
			m.until = until
			return nil
		} else {
			for _, node := range m.nodes {
				if node == nil {
					continue
				}

				reply := node.Cmd("eval", `
					if redis.call("get", KEYS[1]) == ARGV[1] then
					    return redis.call("del", KEYS[1])
					else
					    return 0
					end
				`, 1, m.Name, value)
				if reply.Err != nil {
					continue
				}
			}
		}

		delay := m.Delay
		if delay == 0 {
			delay = DefaultDelay
		}
		time.Sleep(delay)
	}

	return ErrFailed
}

// Unlock unlocks m.
// It is a run-time error if m is not locked on entry to Unlock.
func (m *Mutex) Unlock() {
	m.nodem.Lock()
	defer m.nodem.Unlock()

	value := m.value
	if value == "" {
		panic("redsync: unlock of unlocked mutex")
	}

	m.value = ""
	m.until = time.Unix(0, 0)

	for _, node := range m.nodes {
		if node == nil {
			continue
		}

		node.Cmd("eval", `
			if redis.call("get", KEYS[1]) == ARGV[1] then
			    return redis.call("del", KEYS[1])
			else
			    return 0
			end
		`, 1, m.Name, value)
	}
}
