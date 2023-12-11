package conn

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"rocket/log"
	"sync"

	vhost "github.com/inconshreveable/go-vhost"
)

type Conn interface {
	net.Conn
	log.Logger
	Id() string
	SetType(string)
	CloseRead() error
}

type loggedConn struct {
	tcp *net.TCPConn
	net.Conn
	log.Logger
	id  int32
	typ string
}

type Listener struct {
	net.Addr
	Conns chan *loggedConn
}

// wrapConn wraps a net.Conn with a loggedConn to add logging and identification.
// It detects and handles existing loggedConns to avoid double wrapping.
func wrapConn(conn net.Conn, typ string) *loggedConn {
	switch c := conn.(type) {
	case *vhost.HTTPConn:
		wrapped := c.Conn.(*loggedConn)
		return &loggedConn{wrapped.tcp, conn, wrapped.Logger, wrapped.id, wrapped.typ}
	case *loggedConn:
		return c
	case *net.TCPConn:
		wrapped := &loggedConn{c, conn, log.NewPrefixLogger(), rand.Int31(), typ}
		wrapped.AddLogPrefix(wrapped.Id())
		return wrapped
	}

	return nil
}

// Listen creates a TCP listener for the given address and type.
// It returns a Listener that handles wrapping accepted connections
// and sending them on a channel. TLS is configured if tlsCfg is not nil.
// The listener runs in a goroutine to continuously accept connections.
func Listen(addr, typ string, tlsCfg *tls.Config) (l *Listener, err error) {
	// listen for incoming connections
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}

	l = &Listener{
		Addr:  listener.Addr(),
		Conns: make(chan *loggedConn),
	}

	go func() {
		for {
			rawConn, err := listener.Accept()
			if err != nil {
				log.Error("failed to accept new TCP connection of type %s: %v", typ, err)
				continue
			}

			c := wrapConn(rawConn, typ)
			if tlsCfg != nil {
				c.Conn = tls.Server(c.Conn, tlsCfg)
			}
			c.Info("New connection from %v", c.RemoteAddr())
			l.Conns <- c
		}
	}()
	return
}

// Wrap wraps a net.Conn with a loggedConn to add logging and
// identification.
func Wrap(conn net.Conn, typ string) *loggedConn {
	return wrapConn(conn, typ)
}

// Dial establishes a new TCP connection to the given address.
// It handles wrapping the raw TCP connection, logging, and optionally
// enabling TLS. The connection type parameter typ provides context
// for logging.
func Dial(addr, typ string, tlsCfg *tls.Config) (conn *loggedConn, err error) {
	var rawConn net.Conn
	if rawConn, err = net.Dial("tcp", addr); err != nil {
		return
	}

	conn = wrapConn(rawConn, typ)
	conn.Debug("New connection to: %v", rawConn.RemoteAddr())

	if tlsCfg != nil {
		conn.StartTLS(tlsCfg)
	}

	return
}

// DialHttpProxy dials a connection through an HTTP proxy server.
// It handles connecting to the proxy, sending a CONNECT request,
// and upgrading to TLS if needed.
func DialHttpProxy(proxyUrl, addr, typ string, tlsCfg *tls.Config) (conn *loggedConn, err error) {
	// parse the proxy address
	var parsedUrl *url.URL
	if parsedUrl, err = url.Parse(proxyUrl); err != nil {
		return
	}

	var proxyAuth string
	if parsedUrl.User != nil {
		proxyAuth = "Basic " + base64.StdEncoding.EncodeToString([]byte(parsedUrl.User.String()))
	}

	var proxyTlsConfig *tls.Config
	switch parsedUrl.Scheme {
	case "http":
		proxyTlsConfig = nil
	case "https":
		proxyTlsConfig = new(tls.Config)
	default:
		err = fmt.Errorf("proxy URL scheme must be http or https, got: %s", parsedUrl.Scheme)
		return
	}

	// dial the proxy
	if conn, err = Dial(parsedUrl.Host, typ, proxyTlsConfig); err != nil {
		return
	}

	// send an HTTP proxy CONNECT message
	req, err := http.NewRequest("CONNECT", "https://"+addr, nil)
	if err != nil {
		return
	}

	if proxyAuth != "" {
		req.Header.Set("Proxy-Authorization", proxyAuth)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; rocket)")
	req.Write(conn)

	// read the proxy's response
	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("Non-200 response from proxy server: %s", resp.Status)
		return
	}

	// upgrade to TLS
	conn.StartTLS(tlsCfg)

	return
}

// StartTLS upgrades the connection to TLS.
// It uses the provided tls.Config to initialize a TLS
// client connection wrapped around the existing connection.
func (c *loggedConn) StartTLS(tlsCfg *tls.Config) {
	c.Conn = tls.Client(c.Conn, tlsCfg)
}

// Close closes the underlying connection after logging a debug message.
// It returns any error from closing the underlying connection.
func (c *loggedConn) Close() (err error) {
	if err := c.Conn.Close(); err == nil {
		c.Debug("Closing")
	}
	return
}

// Id returns a unique identifier for the connection composed of the connection
// type and id. This is used in log messages to identify the connection.
func (c *loggedConn) Id() string {
	return fmt.Sprintf("%s:%x", c.typ, c.id)
}

// SetType changes the type identifier used in the connection ID. It updates
// the log prefixes to use the new ID.
func (c *loggedConn) SetType(typ string) {
	oldId := c.Id()
	c.typ = typ
	c.ClearLogPrefixes()
	c.AddLogPrefix(c.Id())
	c.Info("Renamed connection %s", oldId)
}

func (c *loggedConn) CloseRead() error {
	// XXX: use CloseRead() in Conn.Join() and in Control.shutdown() for cleaner
	// connection termination. Unfortunately, when I've tried that, I've observed
	// failures where the connection was closed *before* flushing its write buffer,
	// set with SetLinger() set properly (which it is by default).
	return c.tcp.CloseRead()
}

// Join copies data between two connections bidirectionally until both are closed.
// It returns the number of bytes copied from c1 to c2 and from c2 to c1.
func Join(c Conn, c2 Conn) (int64, int64) {
	var wait sync.WaitGroup

	pipe := func(to Conn, from Conn, bytesCopied *int64) {
		defer to.Close()
		defer from.Close()
		defer wait.Done()

		var err error
		*bytesCopied, err = io.Copy(to, from)
		if err != nil {
			from.Warn("Copied %d bytes to %s before failing with error %v", *bytesCopied, to.Id(), err)
		} else {
			from.Debug("Copied %d bytes to %s", *bytesCopied, to.Id())
		}
	}

	wait.Add(2)
	var fromBytes, toBytes int64
	go pipe(c, c2, &fromBytes)
	go pipe(c2, c, &toBytes)
	c.Info("Joined with connection %s", c2.Id())
	wait.Wait()
	return fromBytes, toBytes
}
