package helpers

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database"
)

// startHandshakeOnlyServer starts a server that completes the Postgres startup handshake on every connection and then
// ignores it. It returns the server address and a function that stops the server.
func startHandshakeOnlyServer() (string, func(), error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to listen: %w", err)
	}

	var mu sync.Mutex
	var conns []net.Conn
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			mu.Lock()
			conns = append(conns, conn)
			mu.Unlock()

			go func() {
				// Read the startup message: a 4 byte length that includes itself, then the body.
				var length int32
				err := binary.Read(conn, binary.BigEndian, &length)
				if err != nil || length < 4 {
					return
				}

				_, err = io.CopyN(io.Discard, conn, int64(length-4))
				if err != nil {
					return
				}

				// Reply with AuthenticationOk and ReadyForQuery (idle).
				_, _ = conn.Write([]byte{'R', 0, 0, 0, 8, 0, 0, 0, 0, 'Z', 0, 0, 0, 5, 'I'})
			}()
		}
	}()

	stop := func() {
		_ = listener.Close()
		mu.Lock()
		for _, conn := range conns {
			_ = conn.Close()
		}

		mu.Unlock()
	}

	return listener.Addr().String(), stop, nil
}

// NewDBWithExhaustedPool returns a database handle whose connection pool only allows one open connection, and that
// connection is already in use. Anything that needs a connection has to wait for one. The returned function releases
// the connection and closes the handle.
func NewDBWithExhaustedPool() (*database.DB, func(), error) {
	address, stopServer, err := startHandshakeOnlyServer()
	if err != nil {
		return nil, nil, err
	}

	db, err := database.NewDB(database.DBConfig{
		DBHost:         address,
		DBUser:         "admin",
		DBName:         "cm",
		DBMaxIdleConns: 1,
		DBMaxOpenConns: 1,
		DBDisableTLS:   true,
	})
	if err != nil {
		stopServer()
		return nil, nil, err
	}

	conn, err := db.Conn().Conn(context.Background())
	if err != nil {
		_ = db.Close()
		stopServer()
		return nil, nil, fmt.Errorf("failed to reserve the only connection: %w", err)
	}

	cleanup := func() {
		_ = conn.Close()
		_ = db.Close()
		stopServer()
	}

	return db, cleanup, nil
}
