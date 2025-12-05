package executor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"yapi.run/cli/internal/domain"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

// GRPCTransport is the transport function for gRPC requests.
func GRPCTransport(ctx context.Context, req *domain.Request) (*domain.Response, error) {
	// Extract metadata
	service := req.Metadata["service"]
	rpc := req.Metadata["rpc"]
	protoFile := req.Metadata["proto"]
	protoPath := req.Metadata["proto_path"]
	insecureFlag, _ := strconv.ParseBool(req.Metadata["insecure"])
	plaintext, _ := strconv.ParseBool(req.Metadata["plaintext"])

	// Connection setup
	target := strings.TrimPrefix(req.URL, "grpc://")
	var opts []grpc.DialOption
	if insecureFlag || plaintext || strings.HasPrefix(target, "localhost") || strings.HasPrefix(target, "127.0.0.1") {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Establish connection
	cc, err := grpc.DialContext(ctx, target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial gRPC target %s: %w", target, err)
	}
	defer cc.Close()

	// Determine descriptor source
	var descSource grpcurl.DescriptorSource
	if protoFile != "" {
		// TODO: Handle proto and proto_path. For now, we focus on reflection.
		_ = protoPath // Avoid unused variable error
		return nil, fmt.Errorf("proto file support not yet implemented")
	} else {
		// Use server reflection
		refClient := grpcreflect.NewClient(ctx, grpc_reflection_v1alpha.NewServerReflectionClient(cc))
		descSource = grpcurl.DescriptorSourceFromServer(ctx, refClient)
	}

	// Prepare request payload
	var reqData []byte
	if req.Body != nil {
		var buf bytes.Buffer
		if _, err = io.Copy(&buf, req.Body); err != nil {
			return nil, fmt.Errorf("failed to read gRPC request body: %w", err)
		}
		reqData = buf.Bytes()
	}

	// Create a RequestSupplier to feed the request data
	reqSupplier := func(m proto.Message) error {
		if len(reqData) == 0 {
			return io.EOF // No more data
		}
		err := (&jsonpb.Unmarshaler{AllowUnknownFields: true}).Unmarshal(bytes.NewReader(reqData), m)
		if err != nil {
			return fmt.Errorf("failed to unmarshal request data: %w", err)
		}
		reqData = nil // Clear data after first use for unary/server-streaming RPCs
		return nil
	}

	// Setup output buffer for handler
	respBuf := bytes.NewBuffer(nil)
	formatter := grpcurl.NewJSONFormatter(true, nil)
	handler := grpcurl.NewDefaultEventHandler(respBuf, descSource, formatter, false)

	// Invoke RPC
	if err := grpcurl.InvokeRPC(ctx, descSource, cc, service+"/"+rpc, nil, handler, reqSupplier); err != nil {
		return nil, fmt.Errorf("failed to invoke gRPC RPC %s/%s: %w", service, rpc, err)
	}

	return &domain.Response{
		StatusCode: 0, // gRPC status is handled differently, 0 for OK
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       io.NopCloser(respBuf),
	}, nil
}
