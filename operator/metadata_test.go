package operator

import (
	"encoding/base64"
	"maps"
	"testing"

	"github.com/nekrassov01/ignoresync/testutil"
)

var testMetadataInput = map[string]string{
	metadataKeySchemeVersion: "1",
	metadataKeyKeyID:         "1",
	metadataKeyCloudKey:      base64.StdEncoding.EncodeToString([]byte("wck")),
	metadataKeyRepoKey:       base64.StdEncoding.EncodeToString([]byte("wrk")),
	metadataKeyDataKey:       base64.StdEncoding.EncodeToString([]byte("wdk")),
	metadataKeyBaseNonce:     base64.StdEncoding.EncodeToString([]byte("baseNonce")),
	metadataKeyChunkSize:     "5242880",
	metadataKeyGitUser:       "alice",
}

var testMetadataOutput = &metadata{
	SchemeVersion:   1,
	KeyID:           "1",
	WrappedCloudKey: []byte("wck"),
	WrappedRepoKey:  []byte("wrk"),
	WrappedDataKey:  []byte("wdk"),
	BaseNonce:       []byte("baseNonce"),
	ChunkSize:       5242880,
	GitUser:         "alice",
}

func Test_parseMetadata(t *testing.T) {
	type args struct {
		m map[string]string
	}
	type want struct {
		metadata *metadata
		isError  bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "success",
			args: args{
				m: testMetadataInput,
			},
			want: want{
				metadata: testMetadataOutput,
				isError:  false,
			},
		},
		{
			name: "missing schemeVersion",
			args: args{
				m: func() map[string]string {
					m := make(map[string]string, len(testMetadataInput))
					maps.Copy(m, testMetadataInput)
					delete(m, metadataKeySchemeVersion)
					return m
				}(),
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
		{
			name: "empty SchemeVersion",
			args: args{
				m: func() map[string]string {
					m := make(map[string]string, len(testMetadataInput))
					maps.Copy(m, testMetadataInput)
					m[metadataKeySchemeVersion] = ""
					return m
				}(),
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
		{
			name: "invalid SchemeVersion",
			args: args{
				m: func() map[string]string {
					m := make(map[string]string, len(testMetadataInput))
					maps.Copy(m, testMetadataInput)
					m[metadataKeySchemeVersion] = "invalid"
					return m
				}(),
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
		{
			name: "missing KeyID",
			args: args{
				m: func() map[string]string {
					m := make(map[string]string, len(testMetadataInput))
					maps.Copy(m, testMetadataInput)
					delete(m, metadataKeyKeyID)
					return m
				}(),
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
		{
			name: "empty KeyID",
			args: args{
				m: func() map[string]string {
					m := make(map[string]string, len(testMetadataInput))
					maps.Copy(m, testMetadataInput)
					m[metadataKeyKeyID] = ""
					return m
				}(),
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
		{
			name: "missing WrappedCloudKey",
			args: args{
				m: func() map[string]string {
					m := make(map[string]string, len(testMetadataInput))
					maps.Copy(m, testMetadataInput)
					delete(m, metadataKeyCloudKey)
					return m
				}(),
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
		{
			name: "empty WrappedCloudKey",
			args: args{
				m: func() map[string]string {
					m := make(map[string]string, len(testMetadataInput))
					maps.Copy(m, testMetadataInput)
					m[metadataKeyCloudKey] = ""
					return m
				}(),
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
		{
			name: "invalid WrappedCloudKey",
			args: args{
				m: func() map[string]string {
					m := make(map[string]string, len(testMetadataInput))
					maps.Copy(m, testMetadataInput)
					m[metadataKeyCloudKey] = "invalid"
					return m
				}(),
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
		{
			name: "missing WrappedRepoKey",
			args: args{
				m: func() map[string]string {
					m := make(map[string]string, len(testMetadataInput))
					maps.Copy(m, testMetadataInput)
					delete(m, metadataKeyRepoKey)
					return m
				}(),
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
		{
			name: "empty WrappedRepoKey",
			args: args{
				m: func() map[string]string {
					m := make(map[string]string, len(testMetadataInput))
					maps.Copy(m, testMetadataInput)
					m[metadataKeyRepoKey] = ""
					return m
				}(),
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
		{
			name: "invalid WrappedRepoKey",
			args: args{
				m: func() map[string]string {
					m := make(map[string]string, len(testMetadataInput))
					maps.Copy(m, testMetadataInput)
					m[metadataKeyRepoKey] = "invalid"
					return m
				}(),
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
		{
			name: "missing WrappedDataKey",
			args: args{
				m: func() map[string]string {
					m := make(map[string]string, len(testMetadataInput))
					maps.Copy(m, testMetadataInput)
					delete(m, metadataKeyDataKey)
					return m
				}(),
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
		{
			name: "empty WrappedDataKey",
			args: args{
				m: func() map[string]string {
					m := make(map[string]string, len(testMetadataInput))
					maps.Copy(m, testMetadataInput)
					m[metadataKeyDataKey] = ""
					return m
				}(),
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
		{
			name: "invalid WrappedDataKey",
			args: args{
				m: func() map[string]string {
					m := make(map[string]string, len(testMetadataInput))
					maps.Copy(m, testMetadataInput)
					m[metadataKeyDataKey] = "invalid"
					return m
				}(),
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
		{
			name: "missing BaseNonce",
			args: args{
				m: func() map[string]string {
					m := make(map[string]string, len(testMetadataInput))
					maps.Copy(m, testMetadataInput)
					delete(m, metadataKeyBaseNonce)
					return m
				}(),
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
		{
			name: "empty BaseNonce",
			args: args{
				m: func() map[string]string {
					m := make(map[string]string, len(testMetadataInput))
					maps.Copy(m, testMetadataInput)
					m[metadataKeyBaseNonce] = ""
					return m
				}(),
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
		{
			name: "invalid BaseNonce",
			args: args{
				m: func() map[string]string {
					m := make(map[string]string, len(testMetadataInput))
					maps.Copy(m, testMetadataInput)
					m[metadataKeyBaseNonce] = "invalid"
					return m
				}(),
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
		{
			name: "missing ChunkSize",
			args: args{
				m: func() map[string]string {
					m := make(map[string]string, len(testMetadataInput))
					maps.Copy(m, testMetadataInput)
					delete(m, metadataKeyChunkSize)
					return m
				}(),
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
		{
			name: "empty ChunkSize",
			args: args{
				m: func() map[string]string {
					m := make(map[string]string, len(testMetadataInput))
					maps.Copy(m, testMetadataInput)
					m[metadataKeyChunkSize] = ""
					return m
				}(),
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
		{
			name: "invalid ChunkSize",
			args: args{
				m: func() map[string]string {
					m := make(map[string]string, len(testMetadataInput))
					maps.Copy(m, testMetadataInput)
					m[metadataKeyChunkSize] = "invalid"
					return m
				}(),
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
		{
			name: "missing GitUser",
			args: args{
				m: func() map[string]string {
					m := make(map[string]string, len(testMetadataInput))
					maps.Copy(m, testMetadataInput)
					delete(m, metadataKeyGitUser)
					return m
				}(),
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
		{
			name: "empty GitUser",
			args: args{
				m: func() map[string]string {
					m := make(map[string]string, len(testMetadataInput))
					maps.Copy(m, testMetadataInput)
					m[metadataKeyGitUser] = ""
					return m
				}(),
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseMetadata(test.args.m)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, got, test.want.metadata)
		})
	}
}
