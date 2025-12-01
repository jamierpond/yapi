package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	"yapi.run/cli/internal/config"
)

// GRPCExecutor handles gRPC requests.
type GRPCExecutor struct{}

// NewGRPCExecutor creates a new GRPCExecutor.
func NewGRPCExecutor() *GRPCExecutor {
	return &GRPCExecutor{}
}

// Execute performs a gRPC request based on the provided YapiConfig.
func (e *GRPCExecutor) Execute(cfg *config.YapiConfig) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // TODO: Make timeout configurable
	defer cancel()

	target := strings.TrimPrefix(cfg.URL, "grpc://")

	var opts []grpc.DialOption
	if cfg.Insecure || cfg.Plaintext {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Establish connection
	cc, err := grpc.DialContext(ctx, target, opts...)
	if err != nil {
		return "", fmt.Errorf("failed to dial gRPC target %s: %w", target, err)
	}
	defer cc.Close()

	// Determine descriptor source
	var descSource grpcurl.DescriptorSource
	if cfg.Proto != "" {
		// TODO: Handle proto and proto_path. For now, we focus on reflection.
		return "", fmt.Errorf("proto file support not yet implemented")
	} else {
		// Use server reflection
		refClient := grpcreflect.NewClient(ctx, grpc_reflection_v1alpha.NewServerReflectionClient(cc))
		descSource = grpcurl.DescriptorSourceFromServer(ctx, refClient)
	}

	// Prepare request payload
	var reqData []byte
	if cfg.Body != nil {
		reqData, err = json.Marshal(cfg.Body)
		if err != nil {
			return "", fmt.Errorf("failed to marshal gRPC request body: %w", err)
		}
	} else if cfg.JSON != "" {
		reqData = []byte(cfg.JSON)
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
	} // Setup output buffer for handler
	respBuf := bytes.NewBuffer(nil)
	// TODO: Handle error output and verbose mode properly
	formatter := grpcurl.NewJSONFormatter(true, nil)
	handler := grpcurl.NewDefaultEventHandler(respBuf, descSource, formatter, false)

	// Invoke RPC
	if err := grpcurl.InvokeRPC(ctx, descSource, cc, cfg.Service+"/"+cfg.RPC, nil, handler, reqSupplier); err != nil {
		return "", fmt.Errorf("failed to invoke gRPC RPC %s/%s: %w", cfg.Service, cfg.RPC, err)
	}

	return respBuf.String(), nil
}
