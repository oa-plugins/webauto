package playwright

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/oa-plugins/webauto/pkg/ipc"
)

var errSessionClosed = errors.New("playwright session worker closed")

type managedSession struct {
	session *Session
	worker  *sessionWorker
}

type commandResult struct {
	resp *ipc.NodeResponse
	err  error
}

type commandRequest struct {
	ctx      context.Context
	payload  map[string]interface{}
	resultCh chan commandResult
}

type sessionWorker struct {
	session *Session
	conn    net.Conn
	reader  *bufio.Reader

	requests chan *commandRequest
	stopCh   chan struct{}
	doneCh   chan struct{}

	closeOnce sync.Once
	closed    atomic.Bool
}

func newSessionWorker(ctx context.Context, session *Session) (*sessionWorker, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	dialCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := (&net.Dialer{}).DialContext(dialCtx, "tcp", fmt.Sprintf("127.0.0.1:%d", session.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to dial session worker: %w", err)
	}

	worker := &sessionWorker{
		session:  session,
		conn:     conn,
		reader:   bufio.NewReader(conn),
		requests: make(chan *commandRequest),
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}

	go worker.loop()

	return worker, nil
}

func (w *sessionWorker) loop() {
	defer close(w.doneCh)

	for {
		select {
		case <-w.stopCh:
			return
		case req, ok := <-w.requests:
			if !ok {
				return
			}
			if req == nil {
				continue
			}
			if w.isClosed() {
				w.deliver(req, nil, errSessionClosed)
				continue
			}
			w.processRequest(req)
		}
	}
}

func (w *sessionWorker) processRequest(req *commandRequest) {
	if err := req.ctx.Err(); err != nil {
		w.deliver(req, nil, err)
		return
	}

	if err := w.writeCommand(req.ctx, req.payload); err != nil {
		w.fail(err)
		w.deliver(req, nil, err)
		return
	}

	resp, err := w.readResponse(req.ctx)
	if err != nil {
		w.fail(err)
		w.deliver(req, nil, err)
		return
	}

	w.deliver(req, resp, nil)
}

func (w *sessionWorker) writeCommand(ctx context.Context, payload map[string]interface{}) error {
	commandJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	deadline := deadlineFromContext(ctx, 5*time.Second)
	if err := w.conn.SetWriteDeadline(deadline); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}

	if _, err := w.conn.Write(append(commandJSON, '\n')); err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	return nil
}

func (w *sessionWorker) readResponse(ctx context.Context) (*ipc.NodeResponse, error) {
	deadline := deadlineFromContext(ctx, 30*time.Second)
	if err := w.conn.SetReadDeadline(deadline); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	line, err := w.reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	line = bytes.TrimSpace(line)
	if len(line) == 0 {
		return nil, fmt.Errorf("received empty response")
	}

	var resp ipc.NodeResponse
	if err := json.Unmarshal(line, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}

func (w *sessionWorker) deliver(req *commandRequest, resp *ipc.NodeResponse, err error) {
	select {
	case req.resultCh <- commandResult{resp: resp, err: err}:
	default:
	}
}

func (w *sessionWorker) Send(ctx context.Context, payload map[string]interface{}) (*ipc.NodeResponse, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if w.isClosed() {
		return nil, errSessionClosed
	}

	req := &commandRequest{
		ctx:      ctx,
		payload:  payload,
		resultCh: make(chan commandResult, 1),
	}

	select {
	case w.requests <- req:
	case <-w.stopCh:
		return nil, errSessionClosed
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	select {
	case result := <-req.resultCh:
		return result.resp, result.err
	case <-w.stopCh:
		return nil, errSessionClosed
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (w *sessionWorker) Close() {
	w.closeOnce.Do(func() {
		w.closed.Store(true)
		close(w.stopCh)
		if w.conn != nil {
			_ = w.conn.Close()
		}
	})

	<-w.doneCh
}

func (w *sessionWorker) fail(err error) {
	var netErr net.Error
	if errors.As(err, &netErr) {
		w.Close()
		return
	}

	if errors.Is(err, net.ErrClosed) {
		w.Close()
		return
	}
}

func (w *sessionWorker) isClosed() bool {
	return w.closed.Load()
}

func deadlineFromContext(ctx context.Context, fallback time.Duration) time.Time {
	if ctx == nil {
		return time.Now().Add(fallback)
	}

	if dl, ok := ctx.Deadline(); ok {
		return dl
	}

	return time.Now().Add(fallback)
}
