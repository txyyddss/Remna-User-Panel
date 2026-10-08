package connectivity

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"net/http"
	"net/http/httptrace"
	"strings"
	"sync/atomic"
	"time"

	xraynet "github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/common/session"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf/serial"
)

// Xray owns global logging hooks, so one process runs one engine lifecycle at a
// time. Waiting is cancellable and counts toward the configured probe timeout.
var xrayLifecycle = make(chan struct{}, 1)

// XrayProbe executes isolated authenticated HTTP checks through embedded Xray.
// Its caller must dispatch checks through the application's external job queue.
type XrayProbe struct {
	// Production uses system roots. The private override supports trusted hosted
	// TLS fixtures without introducing a runtime certificate-bypass setting.
	rootCAs *x509.CertPool
	// The private factory permits startup-failure cleanup coverage in hosted CI.
	newInstance func(context.Context, *core.Config) (*core.Instance, error)
}

// NewXrayProbe returns a stateless probe; credentials exist only during Check.
func NewXrayProbe() Probe { return &XrayProbe{} }

// Check measures TTFB through exactly one proxy without redirects or fallback.
func (probe *XrayProbe) Check(parent context.Context, target Target, config Config) (outcome Outcome) {
	if config.TimeoutSeconds <= 0 || config.TimeoutSeconds > 60 {
		return probeOutcome("error", "INVALID_RESOLVED_CONFIG")
	}
	ctx, cancel := context.WithTimeout(parent, time.Duration(config.TimeoutSeconds)*time.Second)
	defer cancel()
	select {
	case xrayLifecycle <- struct{}{}:
		defer func() { <-xrayLifecycle }()
	case <-ctx.Done():
		return probeRequestError(parent, ctx, ctx.Err())
	}
	if ctx.Err() != nil {
		return probeRequestError(parent, ctx, ctx.Err())
	}
	quietXrayLogging()
	data, err := projectTarget(target)
	if err != nil {
		code := ErrorCode(err)
		status := "error"
		if strings.HasPrefix(code, "CONNECTIVITY_UNSUPPORTED_") {
			status = "unsupported"
		}
		return Outcome{Status: status, ErrorCode: code}
	}
	decoded, err := serial.DecodeJSONConfigStrict(strings.NewReader(string(data)))
	if err != nil {
		return probeOutcome("error", "XRAY_CONFIG_INVALID")
	}
	compiled, err := decoded.Build()
	if err != nil {
		return probeOutcome("error", "XRAY_CONFIG_INVALID")
	}
	if ctx.Err() != nil {
		return probeRequestError(parent, ctx, ctx.Err())
	}
	outboundTag := decoded.OutboundConfigs[0].Tag
	constructor := probe.newInstance
	if constructor == nil {
		constructor = core.NewWithContext
	}
	instance, err := constructor(ctx, compiled)
	if instance != nil {
		defer func() {
			if err := instance.Close(); err != nil && outcome.Status == "connected" {
				outcome = probeOutcome("error", "PROBE_FAILED")
			}
		}()
	}
	if err != nil || instance == nil {
		return probeOutcome("error", "XRAY_START_FAILED")
	}
	if ctx.Err() != nil {
		return probeRequestError(parent, ctx, ctx.Err())
	}
	// StartInstance loses the instance on startup failure in the pinned core.
	// Owning it before Start guarantees teardown and binds features to the probe.
	if err := instance.Start(); err != nil {
		return probeOutcome("error", "XRAY_START_FAILED")
	}
	transport := &http.Transport{
		Proxy: nil, DisableKeepAlives: true, DisableCompression: true,
		TLSClientConfig:        &tls.Config{RootCAs: probe.rootCAs, MinVersion: tls.VersionTLS12},
		MaxResponseHeaderBytes: 16 << 10, ForceAttemptHTTP2: false,
		TLSHandshakeTimeout: time.Duration(config.TimeoutSeconds) * time.Second,
		DialContext: func(_ context.Context, network, address string) (net.Conn, error) {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			destination, err := xraynet.ParseDestination("tcp:" + address)
			if err != nil || network != "tcp" {
				return nil, projectionError("PROBE_FAILED")
			}
			// net/http may detach a dial from request cancellation. Xray's
			// asynchronous protocol handshake must retain the probe deadline.
			dialCtx := session.SetForcedOutboundTagToContext(ctx, outboundTag)
			return core.Dial(dialCtx, instance, destination)
		},
	}
	defer transport.CloseIdleConnections()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, config.ProbeURL, nil)
	if err != nil || request.URL.Host == "" || request.URL.User != nil || request.URL.Scheme != "https" {
		return probeOutcome("error", "INVALID_RESOLVED_CONFIG")
	}
	var firstByte atomic.Int64
	started := time.Now()
	trace := &httptrace.ClientTrace{GotFirstResponseByte: func() { firstByte.CompareAndSwap(0, time.Since(started).Nanoseconds()) }}
	request = request.WithContext(httptrace.WithClientTrace(ctx, trace))
	client := &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	response, err := client.Do(request)
	if err != nil {
		return probeRequestError(parent, ctx, err)
	}
	// Status mode deliberately closes the body without downloading its contents.
	if err := response.Body.Close(); err != nil {
		return probeRequestError(parent, ctx, err)
	}
	if ctx.Err() != nil {
		return probeRequestError(parent, ctx, ctx.Err())
	}
	status := response.StatusCode
	outcome = Outcome{Status: "connected", HTTPStatus: &status}
	if elapsed := firstByte.Load(); elapsed > 0 {
		latency := float64(elapsed) / float64(time.Millisecond)
		outcome.LatencyMS = &latency
	}
	if status < 200 || status >= 300 {
		outcome.Status, outcome.ErrorCode = "failed", "CONNECTIVITY_UNEXPECTED_HTTP_STATUS"
	}
	return outcome
}

func probeOutcome(status, code string) Outcome {
	return Outcome{Status: status, ErrorCode: "CONNECTIVITY_" + code}
}

func probeRequestError(parent, ctx context.Context, err error) Outcome {
	if parent.Err() != nil {
		return probeOutcome("interrupted", "INTERRUPTED")
	}
	var networkError net.Error
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || (errors.As(err, &networkError) && networkError.Timeout()) {
		return probeOutcome("failed", "PROBE_TIMEOUT")
	}
	return probeOutcome("failed", "PROBE_FAILED")
}
