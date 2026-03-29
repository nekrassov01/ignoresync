package operator

import (
	"bytes"
	"io"
	"testing"

	"github.com/nekrassov01/ignoresync/testutil"
)

func Test_encryptBody(t *testing.T) {
	var (
		key           = bytes.Repeat([]byte{0x11}, keySize)
		baseNonce     = bytes.Repeat([]byte{0x22}, baseNonceSize)
		aad           = []byte("aad")
		plaintext     = bytes.Repeat([]byte("a"), chunkSize+128)
		isClosedKey   = false
		isClosedNonce = false
	)
	type args struct {
		body      io.ReadCloser
		key       []byte
		baseNonce []byte
		aad       []byte
		chunkSize int
	}
	type want struct {
		checkFn func(*testing.T, io.ReadCloser, error)
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "success single chunk",
			args: args{
				body:      io.NopCloser(bytes.NewReader([]byte("hello"))),
				key:       key,
				baseNonce: baseNonce,
				aad:       aad,
				chunkSize: chunkSize,
			},
			want: want{
				checkFn: checkReadCloserLength(len([]byte("hello")) + tagSize),
			},
		},
		{
			name: "success multiple chunks",
			args: args{
				body:      io.NopCloser(bytes.NewReader(plaintext)),
				key:       key,
				baseNonce: baseNonce,
				aad:       aad,
				chunkSize: chunkSize,
			},
			want: want{
				checkFn: checkReadCloserLength(len(plaintext) + tagSize*2),
			},
		},
		{
			name: "success empty body",
			args: args{
				body:      io.NopCloser(bytes.NewReader(nil)),
				key:       key,
				baseNonce: baseNonce,
				aad:       aad,
				chunkSize: chunkSize,
			},
			want: want{
				checkFn: checkReadCloserLength(0),
			},
		},
		{
			name: "invalid key size",
			args: args{
				body:      testutil.IsClosedReadCloser{Closed: &isClosedKey},
				key:       []byte("short"),
				baseNonce: baseNonce,
				aad:       aad,
				chunkSize: chunkSize,
			},
			want: want{
				checkFn: checkReadCloserError(&isClosedKey),
			},
		},
		{
			name: "invalid base nonce size",
			args: args{
				body:      testutil.IsClosedReadCloser{Closed: &isClosedNonce},
				key:       key,
				baseNonce: []byte("short"),
				aad:       aad,
				chunkSize: chunkSize,
			},
			want: want{
				checkFn: checkReadCloserError(&isClosedNonce),
			},
		},
		{
			name: "read plaintext error",
			args: args{
				body:      testutil.ReadErrorReadCloser{},
				key:       key,
				baseNonce: baseNonce,
				aad:       aad,
				chunkSize: chunkSize,
			},
			want: want{
				checkFn: checkReadCloserStreamError(),
			},
		},
		{
			name: "close plaintext error",
			args: args{
				body:      testutil.CloseErrorReadCloser{Reader: bytes.NewReader([]byte("hello"))},
				key:       key,
				baseNonce: baseNonce,
				aad:       aad,
				chunkSize: chunkSize,
			},
			want: want{
				checkFn: checkReadCloserStreamError(),
			},
		},
		{
			name: "invalid chunk size zero",
			args: args{
				body:      io.NopCloser(bytes.NewReader([]byte("hello"))),
				key:       key,
				baseNonce: baseNonce,
				aad:       aad,
				chunkSize: 0,
			},
			want: want{
				checkFn: checkReadCloserStreamError(),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := encryptBody(test.args.body, test.args.key, test.args.baseNonce, test.args.aad, test.args.chunkSize)
			test.want.checkFn(t, got, err)
		})
	}
}

func Test_decryptBody(t *testing.T) {
	var (
		key              = bytes.Repeat([]byte{0x11}, keySize)
		baseNonce        = bytes.Repeat([]byte{0x22}, baseNonceSize)
		aad              = []byte("aad")
		plaintext        = bytes.Repeat([]byte("b"), chunkSize+128)
		ciphertextBasic  = mustEncryptBody(t, plaintext, key, baseNonce, aad)
		ciphertextSingle = mustEncryptBody(t, []byte("single"), key, baseNonce, aad)
		isClosedKey      = false
		isClosedNonce    = false
	)
	type args struct {
		body      io.ReadCloser
		key       []byte
		baseNonce []byte
		aad       []byte
		chunkSize int
	}
	type want struct {
		checkFn func(*testing.T, io.ReadCloser, error)
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "success single chunk",
			args: args{
				body:      io.NopCloser(bytes.NewReader(ciphertextSingle)),
				key:       key,
				baseNonce: baseNonce,
				aad:       aad,
				chunkSize: chunkSize,
			},
			want: want{
				checkFn: checkReadCloserValue([]byte("single")),
			},
		},
		{
			name: "success multiple chunks",
			args: args{
				body:      io.NopCloser(bytes.NewReader(ciphertextBasic)),
				key:       key,
				baseNonce: baseNonce,
				aad:       aad,
				chunkSize: chunkSize,
			},
			want: want{
				checkFn: checkReadCloserValue(plaintext),
			},
		},
		{
			name: "success empty body",
			args: args{
				body:      io.NopCloser(bytes.NewReader(nil)),
				key:       key,
				baseNonce: baseNonce,
				aad:       aad,
				chunkSize: chunkSize,
			},
			want: want{
				checkFn: checkReadCloserLength(0),
			},
		},
		{
			name: "invalid key size",
			args: args{
				body:      testutil.IsClosedReadCloser{Closed: &isClosedKey},
				key:       []byte("short"),
				baseNonce: baseNonce,
				aad:       aad,
				chunkSize: chunkSize,
			},
			want: want{
				checkFn: checkReadCloserError(&isClosedKey),
			},
		},
		{
			name: "invalid base nonce size",
			args: args{
				body:      testutil.IsClosedReadCloser{Closed: &isClosedNonce},
				key:       key,
				baseNonce: []byte("short"),
				aad:       aad,
				chunkSize: chunkSize,
			},
			want: want{
				checkFn: checkReadCloserError(&isClosedNonce),
			},
		},
		{
			name: "aad mismatch",
			args: args{
				body:      io.NopCloser(bytes.NewReader(ciphertextBasic)),
				key:       key,
				baseNonce: baseNonce,
				aad:       []byte("other"),
				chunkSize: chunkSize,
			},
			want: want{
				checkFn: checkReadCloserStreamError(),
			},
		},
		{
			name: "ciphertext truncated",
			args: args{
				body:      io.NopCloser(bytes.NewReader([]byte("short"))),
				key:       key,
				baseNonce: baseNonce,
				aad:       aad,
				chunkSize: chunkSize,
			},
			want: want{
				checkFn: checkReadCloserStreamError(),
			},
		},
		{
			name: "read ciphertext error",
			args: args{
				body:      testutil.ReadErrorReadCloser{},
				key:       key,
				baseNonce: baseNonce,
				aad:       aad,
				chunkSize: chunkSize,
			},
			want: want{
				checkFn: checkReadCloserStreamError(),
			},
		},
		{
			name: "close ciphertext error",
			args: args{
				body:      testutil.CloseErrorReadCloser{Reader: bytes.NewReader(ciphertextBasic)},
				key:       key,
				baseNonce: baseNonce,
				aad:       aad,
				chunkSize: chunkSize,
			},
			want: want{
				checkFn: checkReadCloserStreamError(),
			},
		},
		{
			name: "invalid chunk size zero",
			args: args{
				body:      io.NopCloser(bytes.NewReader([]byte("hello"))),
				key:       key,
				baseNonce: baseNonce,
				aad:       aad,
				chunkSize: 0,
			},
			want: want{
				checkFn: checkReadCloserStreamError(),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := decryptBody(test.args.body, test.args.key, test.args.baseNonce, test.args.aad, test.args.chunkSize)
			test.want.checkFn(t, got, err)
		})
	}
}

func Test_encryptKey(t *testing.T) {
	var (
		key       = bytes.Repeat([]byte{0x33}, keySize)
		plaintext = []byte("wrapped-key")
		aad       = []byte("aad")
	)
	type args struct {
		key       []byte
		plaintext []byte
		aad       []byte
	}
	type want struct {
		checkFn func(*testing.T, []byte, error)
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "success empty plaintext",
			args: args{
				key:       key,
				plaintext: nil,
				aad:       aad,
			},
			want: want{
				checkFn: checkBytesLength(chunkNonceSize + tagSize),
			},
		},
		{
			name: "success",
			args: args{
				key:       key,
				plaintext: plaintext,
				aad:       aad,
			},
			want: want{
				checkFn: checkBytesLength(chunkNonceSize + len(plaintext) + tagSize),
			},
		},
		{
			name: "invalid key size",
			args: args{
				key:       []byte("short"),
				plaintext: plaintext,
				aad:       aad,
			},
			want: want{
				checkFn: checkBytesError(),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := encryptKey(test.args.key, test.args.plaintext, test.args.aad)
			test.want.checkFn(t, got, err)
		})
	}
}

func Test_decryptKey(t *testing.T) {
	var (
		key             = bytes.Repeat([]byte{0x33}, keySize)
		plaintext       = []byte("wrapped-key")
		aad             = []byte("aad")
		ciphertextBasic = func() []byte { s, _ := encryptKey(key, plaintext, aad); return s }()
		ciphertextEmpty = func() []byte { s, _ := encryptKey(key, nil, aad); return s }()
	)
	type args struct {
		key        []byte
		ciphertext []byte
		aad        []byte
	}
	type want struct {
		value   []byte
		isError bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "success",
			args: args{
				key:        key,
				ciphertext: ciphertextBasic,
				aad:        aad,
			},
			want: want{
				value: plaintext,
			},
		},
		{
			name: "invalid key size",
			args: args{
				key:        []byte("short"),
				ciphertext: ciphertextBasic,
				aad:        aad,
			},
			want: want{
				isError: true,
			},
		},
		{
			name: "success empty plaintext",
			args: args{
				key:        key,
				ciphertext: ciphertextEmpty,
				aad:        aad,
			},
			want: want{
				value: nil,
			},
		},
		{
			name: "invalid wrapped key size",
			args: args{
				key:        key,
				ciphertext: []byte("short"),
				aad:        aad,
			},
			want: want{
				isError: true,
			},
		},
		{
			name: "aad mismatch",
			args: args{
				key:        key,
				ciphertext: ciphertextBasic,
				aad:        []byte("other"),
			},
			want: want{
				isError: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := decryptKey(test.args.key, test.args.ciphertext, test.args.aad)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}

func Test_generateKey(t *testing.T) {
	type want struct {
		length  int
		isError bool
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "success",
			want: want{
				length: keySize,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := generateKey()
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, len(got), test.want.length)
		})
	}
}

func Test_generateNonce(t *testing.T) {
	type want struct {
		length  int
		isError bool
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "success",
			want: want{
				length: baseNonceSize,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := generateNonce()
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, len(got), test.want.length)
		})
	}
}

func Test_setChunkNonce(t *testing.T) {
	tests := []struct {
		name      string
		baseNonce []byte
		idx       uint32
		want      [chunkNonceSize]byte
	}{
		{
			name:      "zero index",
			baseNonce: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			idx:       0,
			want:      [chunkNonceSize]byte{1, 2, 3, 4, 5, 6, 7, 8, 0, 0, 0, 0},
		},
		{
			name:      "non zero index",
			baseNonce: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			idx:       0x01020304,
			want:      [chunkNonceSize]byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4},
		},
		{
			name:      "max index",
			baseNonce: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			idx:       0xffffffff,
			want:      [chunkNonceSize]byte{1, 2, 3, 4, 5, 6, 7, 8, 255, 255, 255, 255},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var got [chunkNonceSize]byte
			setChunkNonce(&got, test.baseNonce, test.idx)
			testutil.CheckValue(t, got, test.want)
		})
	}
}

func Test_deriveLocalKey(t *testing.T) {
	var (
		mk = bytes.Repeat([]byte{0x44}, keySize)
	)
	type args struct {
		mk []byte
	}
	type want struct {
		length  int
		isError bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "success",
			args: args{
				mk: mk,
			},
			want: want{
				length:  keySize,
				isError: false,
			},
		},
		{
			name: "empty master key",
			args: args{
				mk: nil,
			},
			want: want{
				length:  0,
				isError: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := deriveLocalKey(test.args.mk)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, got == nil, test.want.isError)
			testutil.CheckValue(t, len(got), test.want.length)
		})
	}
}

func Test_deriveWrapKey(t *testing.T) {
	var (
		lk = bytes.Repeat([]byte{0x55}, keySize)
		ck = bytes.Repeat([]byte{0x66}, keySize)
	)
	type args struct {
		lk []byte
		ck []byte
	}
	type want struct {
		length  int
		isError bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "success",
			args: args{
				lk: lk,
				ck: ck,
			},
			want: want{
				length:  keySize,
				isError: false,
			},
		},
		{
			name: "empty local key",
			args: args{
				lk: nil,
				ck: ck,
			},
			want: want{
				length:  0,
				isError: true,
			},
		},
		{
			name: "empty cloud key",
			args: args{
				lk: lk,
				ck: nil,
			},
			want: want{
				length:  0,
				isError: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := deriveWrapKey(test.args.lk, test.args.ck)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, got == nil, test.want.isError)
			testutil.CheckValue(t, len(got), test.want.length)
		})
	}
}

func Test_newGCM(t *testing.T) {
	type args struct {
		key []byte
	}
	type want struct {
		nonceSize int
		overhead  int
		isError   bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "success",
			args: args{
				key: bytes.Repeat([]byte{0x77}, keySize),
			},
			want: want{
				nonceSize: chunkNonceSize,
				overhead:  tagSize,
				isError:   false,
			},
		},
		{
			name: "invalid key size",
			args: args{
				key: []byte("short"),
			},
			want: want{
				nonceSize: 0,
				overhead:  0,
				isError:   true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := newGCM(test.args.key)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, got == nil, test.want.isError)
			if got != nil {
				testutil.CheckValue(t, got.NonceSize(), test.want.nonceSize)
				testutil.CheckValue(t, got.Overhead(), test.want.overhead)
			}
		})
	}
}

func Test_buildAAD(t *testing.T) {
	type args struct {
		prefix    string
		baseNonce []byte
		purpose   string
	}
	type want struct {
		value   []byte
		isError bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "success",
			args: args{
				prefix:    "obj",
				purpose:   "p",
				baseNonce: []byte{1, 2, 3},
			},
			want: want{
				value: func() []byte {
					// build expected: [schemeVersion][len(prefix)][prefix][len(purpose)][purpose][len(baseNonce)][baseNonce]
					b := []byte{}
					b = append(b, []byte{0, 0, 0, 1}...) // schemeVersion
					b = append(b, []byte{0, 0, 0, 3}...) // len("obj")
					b = append(b, []byte("obj")...)
					b = append(b, []byte{0, 0, 0, 1}...) // len("p")
					b = append(b, []byte("p")...)
					b = append(b, []byte{0, 0, 0, 3}...) // len(baseNonce)
					b = append(b, []byte{1, 2, 3}...)
					return b
				}(),
				isError: false,
			},
		},
		{
			name: "zero length prefix",
			args: args{
				prefix:    "",
				purpose:   "p",
				baseNonce: []byte{1, 2, 3},
			},
			want: want{
				value:   nil,
				isError: true,
			},
		},
		{
			name: "zero length base nonce",
			args: args{
				prefix:    "obj",
				purpose:   "p",
				baseNonce: nil,
			},
			want: want{
				value:   nil,
				isError: true,
			},
		},
		{
			name: "zero length purpose",
			args: args{
				prefix:    "obj",
				purpose:   "",
				baseNonce: []byte{1, 2, 3},
			},
			want: want{
				value:   nil,
				isError: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := buildAAD(test.args.prefix, test.args.purpose, test.args.baseNonce)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}

// checkReadCloserLength checks that the ReadCloser returns the expected length and no error.
func checkReadCloserLength(want int) func(*testing.T, io.ReadCloser, error) {
	return func(t *testing.T, got io.ReadCloser, err error) {
		t.Helper()
		testutil.CheckError(t, err != nil, false)
		defer func() {
			_ = got.Close()
		}()
		body, readErr := io.ReadAll(got)
		testutil.CheckError(t, readErr != nil, false)
		testutil.CheckValue(t, len(body), want)
	}
}

// checkReadCloserValue checks that the ReadCloser returns the expected value and no error.
func checkReadCloserValue(want []byte) func(*testing.T, io.ReadCloser, error) {
	return func(t *testing.T, got io.ReadCloser, err error) {
		t.Helper()
		testutil.CheckError(t, err != nil, false)
		defer func() {
			_ = got.Close()
		}()
		body, readErr := io.ReadAll(got)
		testutil.CheckError(t, readErr != nil, false)
		testutil.CheckValue(t, body, want)
	}
}

// checkReadCloserError checks that the error is not nil and the ReadCloser is nil.
// If closed is not nil, it also checks that the ReadCloser was closed.
func checkReadCloserError(closed *bool) func(*testing.T, io.ReadCloser, error) {
	return func(t *testing.T, got io.ReadCloser, err error) {
		t.Helper()
		testutil.CheckError(t, err != nil, true)
		testutil.CheckValue(t, got, io.ReadCloser(nil))
		if closed != nil {
			testutil.CheckValue(t, *closed, true)
		}
	}
}

// checkReadCloserStreamError checks that the error is not nil and the ReadCloser
// returns an error on read or close.
func checkReadCloserStreamError() func(*testing.T, io.ReadCloser, error) {
	return func(t *testing.T, got io.ReadCloser, err error) {
		t.Helper()
		testutil.CheckError(t, err != nil, false)
		defer func() {
			_ = got.Close()
		}()
		_, readErr := io.ReadAll(got)
		testutil.CheckError(t, readErr != nil, true)
	}
}

// checkBytesLength checks that the byte slice has the expected length and no error.
func checkBytesLength(want int) func(*testing.T, []byte, error) {
	return func(t *testing.T, got []byte, err error) {
		t.Helper()
		testutil.CheckError(t, err != nil, false)
		testutil.CheckValue(t, len(got), want)
	}
}

// checkBytesError checks that the error is not nil and the byte slice is nil.
func checkBytesError() func(*testing.T, []byte, error) {
	return func(t *testing.T, got []byte, err error) {
		t.Helper()
		testutil.CheckError(t, err != nil, true)
		testutil.CheckValue(t, got, []byte(nil))
	}
}

// mustEncryptBody is a helper function that encrypts the plaintext with the given key, base nonce,
// and aad, and returns the ciphertext. It fails the test if encryption fails.
func mustEncryptBody(t *testing.T, plaintext []byte, key []byte, baseNonce []byte, aad []byte) []byte {
	t.Helper()
	body, err := encryptBody(io.NopCloser(bytes.NewReader(plaintext)), key, baseNonce, aad, chunkSize)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = body.Close()
	}()
	got, err := io.ReadAll(body)
	if err != nil {
		t.Fatal(err)
	}
	return got
}
