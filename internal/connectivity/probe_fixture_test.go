package connectivity

import (
	"crypto/x509"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/proxy/vless"
	"github.com/xtls/xray-core/proxy/vless/encoding"
)

const probeHostUUID = "a1111111-1111-4111-8111-111111111111"
const probeUserUUID = "22222222-2222-4222-8222-222222222222"

func probeResolvedTarget(t *testing.T, proxyAddress, credential string) Target {
	t.Helper()
	address, portText, err := net.SplitHostPort(proxyAddress)
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}
	host := map[string]any{
		"metadata": map[string]any{"uuid": probeHostUUID}, "address": address, "port": port,
		"protocol": "vless", "protocolOptions": map[string]any{"id": credential, "encryption": "none"},
		"transport": "tcp", "transportOptions": map[string]any{}, "security": "none",
		"clientOverrides": map[string]any{"mapper": map[string]any{"xrayJson": []any{}}},
	}
	data, err := json.Marshal(host)
	if err != nil {
		t.Fatal(err)
	}
	return Target{HostUUID: probeHostUUID, Address: address, Port: port, Resolved: data}
}

func mutateProbeTarget(t *testing.T, target Target, change func(map[string]any)) Target {
	t.Helper()
	var host map[string]any
	if err := json.Unmarshal(target.Resolved, &host); err != nil {
		t.Fatal(err)
	}
	change(host)
	data, err := json.Marshal(host)
	if err != nil {
		t.Fatal(err)
	}
	target.Resolved = data
	return target
}

func trustedProbeTarget(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *XrayProbe) {
	t.Helper()
	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)
	roots := x509.NewCertPool()
	roots.AddCert(server.Certificate())
	return server, &XrayProbe{rootCAs: roots}
}

// A protocol-native fixture uses upstream's VLESS decoder but no second Xray
// Server. Authentication precedes any forwarding to the reachable HTTPS target.
func nativeVLESSFixture(t *testing.T, credential, allowedURL string) (string, *atomic.Int64) {
	t.Helper()
	account, err := (&vless.Account{Id: credential, Encryption: "none"}).AsAccount()
	if err != nil {
		t.Fatal(err)
	}
	validator := new(vless.MemoryValidator)
	if err := validator.Add(&protocol.MemoryUser{Account: account}); err != nil {
		t.Fatal(err)
	}
	target, err := url.Parse(allowedURL)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	accepted := new(atomic.Int64)
	var workers sync.WaitGroup
	workers.Add(1)
	go func() {
		defer workers.Done()
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			workers.Add(1)
			go func() {
				defer workers.Done()
				defer conn.Close() // Fixture teardown; no reusable connection remains.
				if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
					return
				}
				_, request, _, _, err := encoding.DecodeRequestHeader(false, nil, conn, validator)
				if err != nil || request.Command != protocol.RequestCommandTCP || request.Destination().NetAddr() != target.Host {
					return
				}
				accepted.Add(1)
				remote, err := net.DialTimeout("tcp", target.Host, time.Second)
				if err != nil {
					return
				}
				defer remote.Close() // Fixture teardown; closed after either copy ends.
				if err := remote.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
					return
				}
				if err := encoding.EncodeResponseHeader(conn, request, &encoding.Addons{}); err != nil {
					return
				}
				copied := make(chan struct{}, 1)
				go func() {
					defer remote.Close() // Propagate cancellation to the HTTPS request.
					// EOF/reset are expected when the status-only client closes its body.
					_, _ = io.Copy(remote, conn)
					copied <- struct{}{}
				}()
				_, _ = io.Copy(conn, remote)
				conn.Close()   // Unblock the request-copy worker before joining it.
				remote.Close() // The first close may race normal EOF; teardown is idempotent.
				<-copied
			}()
		}
	}()
	t.Cleanup(func() {
		listener.Close() // Accept returns; bounded connection deadlines finish workers.
		workers.Wait()
	})
	return listener.Addr().String(), accepted
}
