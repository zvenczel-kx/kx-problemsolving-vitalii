package registry

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestRegister(t *testing.T) {
	r := NewServiceRegistry(1 * time.Second)

	r.Register("srv1")

	st := r.Status()
	if len(st) != 1 {
		t.Fatalf("expected 1 service, got %d", len(st))
	}
	if st[0].Addr != "srv1" {
		t.Fatalf("expected addr srv1, got %s", st[0].Addr)
	}
	if !st[0].Alive {
		t.Fatalf("expected srv1 to be alive right after register")
	}
}

func TestHeartbeat(t *testing.T) {
	r := NewServiceRegistry(50 * time.Millisecond)

	r.Register("srv1")
	time.Sleep(30 * time.Millisecond)

	r.Heartbeat("srv1")

	time.Sleep(30 * time.Millisecond)

	st := r.Status()
	if len(st) != 1 {
		t.Fatalf("expected 1 service")
	}
	if !st[0].Alive {
		t.Fatalf("expected srv1 to stay alive after heartbeat")
	}
}

func TestStatusTTL(t *testing.T) {
	r := NewServiceRegistry(20 * time.Millisecond)

	r.Register("a")
	r.Register("b")

	time.Sleep(30 * time.Millisecond)

	st := r.Status()
	if st[0].Alive || st[1].Alive {
		t.Fatalf("expected all services to be dead after ttl")
	}
}

func TestStatusSorted(t *testing.T) {
	r := NewServiceRegistry(1 * time.Second)
	r.Register("zeta")
	r.Register("alpha")
	r.Register("beta")

	st := r.Status()

	if st[0].Addr != "alpha" || st[1].Addr != "beta" || st[2].Addr != "zeta" {
		t.Fatalf("services must be sorted by address: %v", st)
	}
}

func TestNextInstanceRoundRobin(t *testing.T) {
	r := NewServiceRegistry(1 * time.Second)

	r.Register("a")
	r.Register("b")
	r.Register("c")

	if r.NextInstance() != "a" {
		t.Fatal("expected first: a")
	}
	if r.NextInstance() != "b" {
		t.Fatal("expected second: b")
	}
	if r.NextInstance() != "c" {
		t.Fatal("expected third: c")
	}
	if r.NextInstance() != "a" {
		t.Fatal("expected wrap-around: a")
	}
}

func TestNextInstanceOnlyAlive(t *testing.T) {
	r := NewServiceRegistry(10 * time.Millisecond)

	r.Register("a")
	r.Register("b")

	if r.NextInstance() != "a" {
		t.Fatal("expected a")
	}

	time.Sleep(12 * time.Millisecond)

	if r.NextInstance() != "" {
		t.Fatal("expected empty because all services are dead")
	}

	r.Register("c")
	if r.NextInstance() != "c" {
		t.Fatal("expected alive c")
	}
}

func TestNextInstanceIndexReset(t *testing.T) {
	r := NewServiceRegistry(1 * time.Second)

	r.Register("a")
	r.Register("b")

	_ = r.NextInstance() // a
	_ = r.NextInstance() // b

	time.Sleep(30 * time.Millisecond)
	r.ttl = 10 * time.Millisecond
	st := r.Status()
	if st[0].Alive || st[1].Alive {
		t.Fatalf("expected both expired")
	}

	r.Register("x")

	if r.NextInstance() != "x" {
		t.Fatal("expected x after index reset")
	}
}

func TestConcurrentRegister(t *testing.T) {
	r := NewServiceRegistry(1 * time.Second)

	const workers = 50
	const perWorker = 100

	var wg sync.WaitGroup
	wg.Add(workers)

	for w := 0; w < workers; w++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < perWorker; i++ {
				addr := fmt.Sprintf("svc-%d-%d", id, i)
				r.Register(addr)
			}
		}(w)
	}

	wg.Wait()

	st := r.Status()
	if len(st) != workers*perWorker {
		t.Fatalf("expected %d services, got %d", workers*perWorker, len(st))
	}
}

func TestConcurrentHeartbeat(t *testing.T) {
	r := NewServiceRegistry(200 * time.Millisecond)

	for i := 0; i < 1000; i++ {
		r.Register(fmt.Sprintf("svc-%d", i))
	}

	const workers = 80
	var wg sync.WaitGroup
	wg.Add(workers)

	for w := 0; w < workers; w++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				addr := fmt.Sprintf("svc-%d", rand.Intn(1000))
				r.Heartbeat(addr)
			}
		}(w)
	}

	wg.Wait()

	st := r.Status()
	if len(st) != 1000 {
		t.Fatalf("expected 1000 services, got %d", len(st))
	}

	for _, s := range st {
		if !s.Alive {
			t.Fatalf("service %s unexpectedly marked dead after concurrent heartbeat", s.Addr)
		}
	}
}

func TestConcurrentRegisterAndHeartbeat(t *testing.T) {
	r := NewServiceRegistry(300 * time.Millisecond)

	const registers = 40
	const heartbeats = 40

	var wg sync.WaitGroup
	wg.Add(registers + heartbeats)

	for i := 0; i < registers; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				r.Register(fmt.Sprintf("reg-%d-%d", id, j))
			}
		}(i)
	}

	for i := 0; i < heartbeats; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				addr := fmt.Sprintf("reg-%d-%d", rand.Intn(registers), rand.Intn(200))
				r.Heartbeat(addr)
			}
		}(i)
	}

	wg.Wait()

	st := r.Status()
	if len(st) != registers*200 {
		t.Fatalf("expected %d services, got %d", registers*200, len(st))
	}

	for _, s := range st {
		if !s.Alive {
			t.Fatalf("some services unexpectedly dead")
		}
	}
}
