package json_test

import (
	"testing"

	"github.com/uber/tchannel-go"
	"github.com/uber/tchannel-go/json"

	"golang.org/x/net/context"

	"github.com/uber/tchannel-go/testutils/testtracing"
)

// JSONHandler tests tracing over JSON encoding
type JSONHandler struct {
	testtracing.TraceHandler
	t *testing.T
}

func (h *JSONHandler) firstCall(ctx context.Context, req *testtracing.TracingRequest) (*testtracing.TracingResponse, error) {
	jctx := json.Wrap(ctx)
	response := new(testtracing.TracingResponse)
	peer := h.Ch.Peers().GetOrAdd(h.Ch.PeerInfo().HostPort)
	if err := json.CallPeer(jctx, peer, h.Ch.PeerInfo().ServiceName, "call", req, response); err != nil {
		return nil, err
	}
	return response, nil
}

func (h *JSONHandler) callJSON(ctx json.Context, req *testtracing.TracingRequest) (*testtracing.TracingResponse, error) {
	return h.HandleCall(ctx, req,
		func(ctx context.Context, req *testtracing.TracingRequest) (*testtracing.TracingResponse, error) {
			jctx := ctx.(json.Context)
			peer := h.Ch.Peers().GetOrAdd(h.Ch.PeerInfo().HostPort)
			childResp := new(testtracing.TracingResponse)
			if err := json.CallPeer(jctx, peer, h.Ch.PeerInfo().ServiceName, "call", req, childResp); err != nil {
				return nil, err
			}
			return childResp, nil
		})
}

func (h *JSONHandler) onError(ctx context.Context, err error) { h.t.Errorf("onError %v", err) }

func TestJSONTracingPropagation(t *testing.T) {
	suite := &testtracing.PropagationTestSuite{
		Encoding: testtracing.EncodingInfo{Format: tchannel.JSON, HeadersSupported: true},
		Register: func(t *testing.T, ch *tchannel.Channel) testtracing.TracingCall {
			handler := &JSONHandler{TraceHandler: testtracing.TraceHandler{Ch: ch}, t: t}
			json.Register(ch, json.Handlers{"call": handler.callJSON}, handler.onError)
			return handler.firstCall
		},
		TestCases: map[testtracing.TracerType][]testtracing.PropagationTestCase{
			testtracing.Noop: {
				{ForwardCount: 2, TracingDisabled: true, ExpectedBaggage: "", ExpectedSpanCount: 0},
				{ForwardCount: 2, TracingDisabled: false, ExpectedBaggage: "", ExpectedSpanCount: 0},
			},
			testtracing.Mock: {
				{ForwardCount: 2, TracingDisabled: true, ExpectedBaggage: testtracing.BaggageValue, ExpectedSpanCount: 0},
				{ForwardCount: 2, TracingDisabled: false, ExpectedBaggage: testtracing.BaggageValue, ExpectedSpanCount: 6},
			},
			testtracing.Jaeger: {
				{ForwardCount: 2, TracingDisabled: true, ExpectedBaggage: testtracing.BaggageValue, ExpectedSpanCount: 0},
				{ForwardCount: 2, TracingDisabled: false, ExpectedBaggage: testtracing.BaggageValue, ExpectedSpanCount: 6},
			},
		},
	}
	suite.Run(t)
}
