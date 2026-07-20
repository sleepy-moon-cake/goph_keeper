package session

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"
)

type Session struct {
	Method     string `json:"method"`
	Token      string `json:"token"`
	UserName   string `json:"user_name"`
	CryptedKey string `json:"key"`
	ExpiredAt  string `json:"expires_at"`
}

type SessionManager struct {
	session Session
	addr    string
	mu      sync.Mutex
}

func NewSessionManager(addr string) *SessionManager {
	return &SessionManager{
		session: Session{},
		addr:    addr,
	}
}

type SessionResponse struct {
	Session *Session `json:"session"`
}

func (s *SessionManager) Listen(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen session: %w", err)
	}

	go func() {
		<-ctx.Done()
		slog.Info("stopping session server due to context cancellation")
		listener.Close()
	}()

	slog.Info("session server started", "addr", s.addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				slog.Warn("accept connection failed", "error", err)
				continue
			}
		}

		go func() {
			defer conn.Close()

			var req Session
			if err := json.NewDecoder(conn).Decode(&req); err != nil {
				slog.Warn("failed to decode request json", "err", err)
				conn.Write([]byte(`{"status":"bad request"}`))
				return
			}

			switch req.Method {
			case "set":
				s.set(req.UserName, req.CryptedKey, req.Token)
				if _, err := conn.Write([]byte(`{"status":"ok"}`)); err != nil {
					slog.Warn("failed to write response for set", "err", err)
				}

			case "get":
				session, found := s.get()
				if !found {
					if _, err := conn.Write([]byte(`{"status":"not found"}`)); err != nil {
						slog.Warn("failed to write response for get (not found)", "err", err)
					}
					return
				}

				response := SessionResponse{Session: session}
				if err := json.NewEncoder(conn).Encode(response); err != nil {
					slog.Warn("failed to encode session response", "err", err)
				}

			case "clear":
				s.clear()
				if _, err := conn.Write([]byte(`{"status":"ok"}`)); err != nil {
					slog.Warn("failed to write response for clear", "err", err)
				}

			case "ping":
				if _, err := conn.Write([]byte(`{"status":"pong"}`)); err != nil {
					slog.Warn("failed to write response for ping", "err", err)
				}

			default:
				slog.Warn("unknown method received", "method", req.Method)
				conn.Write([]byte(`{"status":"unknown method"}`))
			}
		}()
	}
}

func (s *SessionManager) set(name string, cryptedKey string, token string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.session.UserName = name
	s.session.CryptedKey = cryptedKey
	s.session.Token = token
}

func (s *SessionManager) clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.session.UserName = ""
	s.session.CryptedKey = ""
	s.session.Token = ""
}

func (s *SessionManager) get() (*Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.session.UserName == "" || s.session.Token == "" || s.session.CryptedKey == "" {
		return nil, false
	}

	// exclude rase when session is encoding and rapidly calling of set
	copiedSession := s.session
	return &copiedSession, true
}

func NewClientSession(addr string) *ClientSession {
	return &ClientSession{
		addr: addr,
	}
}

type ClientSession struct {
	addr string
}

func (c *ClientSession) Ping(ctx context.Context) error {
	dialer := net.Dialer{
		Timeout: 200 * time.Millisecond,
	}

	conn, err := dialer.DialContext(ctx, "tcp", c.addr)

	if err != nil {
		return fmt.Errorf("Server is not started. Call firstly: start")
	}
	conn.Close()
	return nil
}

func (c *ClientSession) Set(name string, key string, token string) error {
	var session = Session{
		Method:     "set",
		UserName:   name,
		CryptedKey: key,
		Token:      token,
	}

	conn, err := net.DialTimeout("tcp", c.addr, 200*time.Millisecond)
	if err != nil {
		return fmt.Errorf("Server is not started. Call firstly: start")
	}

	if err := json.NewEncoder(conn).Encode(session); err != nil {
		slog.Error("set session: encoding error", "err", err)
		return err
	}

	buffer := make([]byte, 128)
	_, err = conn.Read(buffer)
	if err != nil {
		slog.Error("set session: failed to get response from server", "err", err)
		return err
	}

	return nil
}

func (c *ClientSession) Get() (*Session, error) {
	var sessionReq = Session{
		Method: "get",
	}

	conn, err := net.DialTimeout("tcp", c.addr, 200*time.Millisecond)
	if err != nil {
		return nil, fmt.Errorf("Server is not started. Call firstly: start")
	}
	defer conn.Close()

	if err := json.NewEncoder(conn).Encode(sessionReq); err != nil {
		slog.Error("get session: encoding error", "err", err)
		return nil, err
	}

	var resp SessionResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		slog.Error("get session: decoding response error", "err", err)
		return nil, err
	}

	if resp.Session == nil {
		return nil, fmt.Errorf("session not found")
	}

	return resp.Session, nil
}
