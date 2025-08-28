package xray

import (
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/session"
)

type conn struct {
	session.Conn
}

func (c *conn) ClientCacheEnabled() bool {
	return false
}

func (c *conn) Unwrap() session.Conn {
	return c
}

type listener struct {
	l server.Listener
}

func (l *listener) Accept() (session.Conn, error) {
	c, err := l.l.Accept()
	if err != nil {
		return nil, err
	}

	return WrapConnection(c), nil
}

func (l *listener) Disconnect(conn session.Conn, reason string) error {
	return l.l.Disconnect(conn, reason)
}

func (l *listener) Close() error {
	return l.l.Close()
}

// WrapListeners wraps listener for the df.
func WrapListeners(listeners []func(conf server.Config) (server.Listener, error)) []func(conf server.Config) (server.Listener, error) {
	var wrapped []func(conf server.Config) (server.Listener, error)
	for _, l := range listeners {
		wrapped = append(wrapped, func(conf server.Config) (server.Listener, error) {
			orig, err := l(conf)
			if err != nil {
				return nil, err
			}
			return &listener{l: orig}, nil
		})
	}
	return wrapped
}

// WrapConnection wraps connection to ensure client cache disabled.
func WrapConnection(c session.Conn) session.Conn {
	return &conn{c}
}
