package mobile

import (
	"errors"
	"fmt"
	"sync"

	"github.com/xjasonlyu/tun2socks/v2/core"
	"github.com/xjasonlyu/tun2socks/v2/core/device"
	"github.com/xjasonlyu/tun2socks/v2/core/device/fdbased"
	"github.com/xjasonlyu/tun2socks/v2/proxy/socks"
)

var (
	mu      sync.Mutex
	running bool
	dev     device.Device
)

// StartTun2Socks memulai tunneling gVisor netstack ke proxy SOCKS5 lokal
// fd: file descriptor integer dari VpnService Android
// socksAddr: contoh "127.0.0.1:1080"
// mtu: standar 1500
func StartTun2Socks(fd int, socksAddr string, mtu int) error {
	mu.Lock()
	defer mu.Unlock()

	if running {
		return errors.New("tun2socks is already running")
	}

	tunDev, err := fdbased.Open(uintptr(fd), uint32(mtu))
	if err != nil {
		return fmt.Errorf("failed to open fd: %w", err)
	}
	dev = tunDev

	handler, err := socks.New(socksAddr, "", "")
	if err != nil {
		_ = dev.Close()
		return fmt.Errorf("failed to create socks handler: %w", err)
	}

	err = core.Start(
		core.WithDevice(dev),
		core.WithProxyHandler(handler),
		core.WithStack("gvisor"),
	)
	if err != nil {
		_ = dev.Close()
		return fmt.Errorf("failed to start core: %w", err)
	}

	running = true
	return nil
}

// StopTun2Socks mematikan engine
func StopTun2Socks() {
	mu.Lock()
	defer mu.Unlock()

	if !running {
		return
	}

	if dev != nil {
		_ = dev.Close()
		dev = nil
	}
	core.Stop()
	running = false
}

// IsRunning cek apakah engine aktif
func IsRunning() bool {
	mu.Lock()
	defer mu.Unlock()
	return running
}

