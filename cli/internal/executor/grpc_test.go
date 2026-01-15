package executor

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateDescriptorSourceFromProto(t *testing.T) {
	testdataDir, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatalf("failed to get testdata path: %v", err)
	}

	tests := []struct {
		name           string
		protoFile      string
		protoPath      string
		configFilePath string
		wantErr        bool
		errContains    string
	}{
		{
			name:           "absolute proto path",
			protoFile:      filepath.Join(testdataDir, "test.proto"),
			protoPath:      "",
			configFilePath: "",
			wantErr:        false,
		},
		{
			name:           "relative proto path with config file",
			protoFile:      "testdata/test.proto",
			protoPath:      "",
			configFilePath: filepath.Join(testdataDir, "..", "fake.yapi.yml"),
			wantErr:        false,
		},
		{
			name:           "with proto_path",
			protoFile:      filepath.Join(testdataDir, "test.proto"),
			protoPath:      testdataDir,
			configFilePath: "",
			wantErr:        false,
		},
		{
			name:           "missing proto file",
			protoFile:      "/nonexistent/path/missing.proto",
			protoPath:      "",
			configFilePath: "",
			wantErr:        true,
			errContains:    "proto file not found",
		},
		{
			name:           "relative path without config file falls back to cwd-relative",
			protoFile:      "testdata/test.proto",
			protoPath:      "",
			configFilePath: "",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			descSource, err := createDescriptorSourceFromProto(tt.protoFile, tt.protoPath, tt.configFilePath)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errContains)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if descSource == nil {
				t.Errorf("expected non-nil descriptor source")
			}
		})
	}
}
