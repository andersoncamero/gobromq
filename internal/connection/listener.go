package connection

import (
	"context"
	"fmt"
	"net"
	"sync"
)

type Listener struct {
	Host        string
	Port        int
	Listen      net.Listener
	Connections chan net.Conn
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	Running     bool
	mu          sync.RWMutex
}

func NewListener(host string, port int) *Listener {
	ctx, cancel := context.WithCancel(context.Background())

	return &Listener{
		Host:        host,
		Port:        port,
		Connections: make(chan net.Conn, 100),
		ctx:         ctx,
		cancel:      cancel,
		Running:     false,
	}
}

func (l *Listener) Start() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.Running {
		return fmt.Errorf("listener is already running")
	}

	addr := fmt.Sprintf("%s:%d", l.Host, l.Port)
	listener, err := net.Listen("tcp", addr)

	if err != nil {
		return fmt.Errorf("failed to start listener: %s: %w", addr, err)
	}

	l.Listen = listener
	l.Running = true

	l.wg.Add(1)

	go l.acceptLoop()

	fmt.Printf("Listener started on %s\n", addr)

	return nil
}

func (l *Listener) Stop() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.Running {
		return fmt.Errorf("listener is not running")
	}

	fmt.Println("Stopping listener...")

	l.cancel()

	if l.Listen != nil {
		l.Listen.Close()
	}

	close(l.Connections)

	l.wg.Wait()

	l.Running = false

	fmt.Println("Listener stopped.")

	return nil

}

func (l *Listener) GetConnections() <-chan net.Conn {
	return l.Connections
}

func (l *Listener) IsRunning() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.Running
}

func (l *Listener) acceptLoop() {
	defer l.wg.Done()

	for {
		select {
		case <-l.ctx.Done():
			return
		default:
			conn, err := l.Listen.Accept()
			if err != nil {
				select {
				case <-l.ctx.Done():
					return
				default:
					fmt.Printf("Error accepting connection: %v\n", err)
					continue
				}
			}

			if tcpConn, ok := conn.(*net.TCPConn); ok {
				tcpConn.SetKeepAlive(true)
				tcpConn.SetNoDelay(true)
			}

			select {
			case l.Connections <- conn:
				fmt.Printf("New connection accepted: %v\n", conn.RemoteAddr())
			default:
				fmt.Printf("Connection buffer full: %v\n", conn.RemoteAddr())
				conn.Close()
			}
		}
	}
}
