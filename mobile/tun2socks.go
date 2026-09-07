package mobile

import (
	"errors"
	"fmt"
	"sync"

	"github.com/xjasonlyu/tun2socks/v2/core/device/fdbased"
	"github.com/xjasonlyu/tun2socks/v2/engine"
)

var (
	mu      sync.Mutex
	running bool
	key     *engine.Key
)

// StartTun2Socks menjalankan engine gVisor menggunakan file descriptor dari Android VpnService
func StartTun2Socks(fd int, socksAddr string, mtu int) error {
	mu.Lock()
	defer mu.Unlock()

	if running {
		return errors.New("tun2socks is already running")
	}

	// Buka interface TUN Android
	dev, err := fdbased.Open("tun", uint32(mtu), fd)
	if err != nil {
		return fmt.Errorf("failed to open fd: %w", err)
	}

	// Konfigurasi engine tun2socks v2
	k := &engine.Key{
		Device: dev,
		Proxy:  fmt.Sprintf("socks5://%s", socksAddr),
		Stack:  "gvisor",
		MTU:    mtu,
	}

	// Start engine
	engine.Insert(k)
	if err := engine.Start(); err != nil {
		engine.Stop()
		return fmt.Errorf("failed to start engine: %w", err)
	}

	key = k
	running = true
	return nil
}

// StopTun2Socks mematikan tunnel
func StopTun2Socks() {
	mu.Lock()
	defer mu.Unlock()

	if !running {
		return
	}

	engine.Stop()
	key = nil
	running = false
}

// IsRunning cek status service
func IsRunning() bool {
	mu.Lock()
	defer mu.Unlock()
	return running
}
