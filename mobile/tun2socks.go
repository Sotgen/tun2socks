package mobile

import (
	"errors"
	"fmt"
	"sync"

	"github.com/xjasonlyu/tun2socks/v2/core"
	"github.com/xjasonlyu/tun2socks/v2/core/device"
	"github.com/xjasonlyu/tun2socks/v2/core/device/fdbased"
	_ "github.com/xjasonlyu/tun2socks/v2/proxy"
)

var (
	mu      sync.Mutex
	running bool
	dev     device.Device
)

func StartTun2Socks(fd int, socksAddr string, mtu int) error {
	mu.Lock()
	defer mu.Unlock()

	if running {
		return errors.New("tun2socks is already running")
	}

	tunDev, err := fdbased.Open("tun", uint32(mtu), fd)
	if err != nil {
		return fmt.Errorf("failed to open fd: %w", err)
	}
	dev = tunDev

	proxyURL := fmt.Sprintf("socks5://%s", socksAddr)

	core.RegisterOutputDevice(dev)

	if err := core.Start(core.WithProxy(proxyURL), core.WithStack("gvisor")); err != nil {
		dev.Close()
		dev = nil
		return fmt.Errorf("failed to start core: %w", err)
	}

	running = true
	return nil
}

func StopTun2Socks() {
	mu.Lock()
	defer mu.Unlock()

	if !running {
		return
	}

	core.Stop()

	if dev != nil {
		dev.Close()
		dev = nil
	}

	running = false
}

func IsRunning() bool {
	mu.Lock()
	defer mu.Unlock()
	return running
}
