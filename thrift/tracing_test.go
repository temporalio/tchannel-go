package thrift_test

import (
	json_encoding "encoding/json"
	"testing"

	"github.com/temporalio/tchannel-go"
	"github.com/temporalio/tchannel-go/testutils/testtracing"
	"github.com/temporalio/tchannel-go/thrift"
	gen "github.com/temporalio/tchannel-go/thrift/gen-go/test"

	"golang.org/x/net/context"
)

// ThriftHandler tests tracing over Thrift encoding
type ThriftHandler struct {
	gen.TChanSimpleService // leave nil so calls to unimplemented methods panic.
	testtracing.TraceHandler

	thriftClient gen.TChanSimpleService
	t            *testing.T
}

func requestFromThrift(req *gen.Data) *testtracing.TracingRequest {
	r := new(testtracing.TracingRequest)
	r.ForwardCount = int(req.I3)
	return r
}

func requestToThrift(r *testtracing.TracingRequest) *gen.Data {
	return &gen.Data{I3: int32(r.ForwardCount)}
}

func responseFromThrift(t *testing.T, res *gen.Data) (*testtracing.TracingResponse, error) {
	var r testtracing.TracingResponse
	if err := json_encoding.Unmarshal([]byte(res.S2), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func responseToThrift(t *testing.T, r *testtracing.TracingResponse) (*gen.Data, error) {
	jsonBytes, err := json_encoding.Marshal(r)
	if err != nil {
		return nil, err
	}
	return &gen.Data{S2: string(jsonBytes)}, nil
}

func (h *ThriftHandler) Call(ctx thrift.Context, arg *gen.Data) (*gen.Data, error) {
	req := requestFromThrift(arg)
	res, err := h.HandleCall(ctx, req,
		func(ctx context.Context, req *testtracing.TracingRequest) (*testtracing.TracingResponse, error) {
			tctx := ctx.(thrift.Context)
			res, err := h.thriftClient.Call(tctx, requestToThrift(req))
			if err != nil {
				return nil, err
			}
			return responseFromThrift(h.t, res)
		})
	if err != nil {
		return nil, err
	}
	return responseToThrift(h.t, res)
}

func (h *ThriftHandler) firstCall(ctx context.Context, req *testtracing.TracingRequest) (*testtracing.TracingResponse, error) {
	tctx := thrift.Wrap(ctx)
	res, err := h.thriftClient.Call(tctx, requestToThrift(req))
	if err != nil {
		return nil, err
	}
	return responseFromThrift(h.t, res)
}

func TestThriftTracingPropagation(t *testing.T) {
	suite := &testtracing.PropagationTestSuite{
		Encoding: testtracing.EncodingInfo{Format: tchannel.Thrift, HeadersSupported: true},
		Register: func(t *testing.T, ch *tchannel.Channel) testtracing.TracingCall {
			opts := &thrift.ClientOptions{HostPort: ch.PeerInfo().HostPort}
			thriftClient := thrift.NewClient(ch, ch.PeerInfo().ServiceName, opts)
			handler := &ThriftHandler{
				TraceHandler: testtracing.TraceHandler{Ch: ch},
				t:            t,
				thriftClient: gen.NewTChanSimpleServiceClient(thriftClient),
			}

			// Register Thrift handler
			server := thrift.NewServer(ch)
			server.Register(gen.NewTChanSimpleServiceServer(handler))

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
