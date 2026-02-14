package binance_ws

import (
	"context"
	"testing"
	"time"
)

func TestReadLoop_NilCtx(t *testing.T) {
	c := NewClient()
	err := c.ReadLoop(context.Background(), func(ctx context.Context, data *StreamData) {})
	if err != ErrNotConnected {
		t.Fatalf("expected ErrNotConnected, got %v", err)
	}
}

func TestWithTestnet_Option(t *testing.T) {
	// Test the Option
	c2 := NewClient(WithTestnet())
	if c2.endpoint != testWsBaseEndpoint {
		t.Errorf("expected endpoint %s, got %s", testWsBaseEndpoint, c2.endpoint)
	}
}

func TestClose(t *testing.T) {
	c := NewClient()
	// Connect requires a real connection or a mock.
	// Since we don't have a mock, we just test that Close() works on uninitialized client.
	err := c.Close()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// dataCh should be nil (in our implementation) or closed
	if c.dataCh != nil {
		t.Error("dataCh should be nil after Close")
	}

	// Double close should not panic
	err = c.Close()
	if err != nil {
		t.Errorf("unexpected error on second close: %v", err)
	}
}

func TestSendDeadlock(t *testing.T) {
	c := NewClient()
	// Эмулируем подключение
	conn := c.newConnection()
	conn.ctx, conn.cancel = context.WithCancel(context.Background())
	conn.writeCh = make(chan writeRequest)

	c.mu.Lock()
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.conns = append(c.conns, conn)
	c.mu.Unlock()

	// Запускаем горутину, которая будет вызывать send и блокироваться на writeCh
	go func() {
		_ = conn.send("TEST", []string{"stream1"})
	}()

	// Даем время горутине заблокироваться
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	select {
	case <-ticker.C:
	}

	// Пытаемся вызвать reconnect, который раньше вызывал дедлок
	done := make(chan struct{})
	go func() {
		conn.reconnect()
		close(done)
	}()

	select {
	case <-done:
		// Успех
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Deadlock detected in send/reconnect")
	}
}

func TestClientReuse(t *testing.T) {
	c := NewClient()

	// Первый коннект (фейковый)
	ctx, cancel := context.WithCancel(context.Background())
	_ = c.Connect(ctx)
	cancel()
	c.Close()

	// Второй коннект
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	err := c.Connect(ctx2)
	if err != nil {
		t.Fatalf("Failed to reuse client: %v", err)
	}

	if c.ctx == nil {
		t.Fatal("ctx should not be nil after Connect")
	}

	c.Close()
}
