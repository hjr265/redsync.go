package redsync_test

import "github.com/hjr265/redsync.go/redsync"

func ExampleMutex() {
	m, err := redsync.NewMutexWithGenericPool("FlyingSquirrels", pools)
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
