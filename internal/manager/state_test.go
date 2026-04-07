package manager

import (
	"bytes"
	"encoding/gob"
	"io"
	"testing"

	"github.com/nekrassov01/ignoresync/internal/params"
	"github.com/nekrassov01/ignoresync/internal/testutil"
	"github.com/zalando/go-keyring"
)

func TestAddCredential(t *testing.T) {
	type fields struct {
		Account string
		Region  string
		OSUser  string
		w       io.Writer
	}
	type args struct {
		id  string
		key []byte
	}
	type want struct {
		isError bool
	}
	type hook struct {
		before func()
		after  func()
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
		hook   hook
	}{
		{
			name: "success",
			fields: fields{
				Account: "012345678901",
				Region:  "us-east-1",
				OSUser:  "os-user",
			},
			args: args{
				id:  "11111111aaaaaaaa",
				key: []byte{48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 48, 49, 50, 51, 52, 53},
			},
			want: want{
				isError: false,
			},
			hook: hook{
				before: func() {
					keyring.MockInit()
					buf := &bytes.Buffer{}
					_ = gob.NewEncoder(buf).Encode(NewMockState())
					_ = keyring.Set(params.CanonicalName, "os-user", buf.String())
				},
				after: func() {
					_ = keyring.Delete(params.CanonicalName, "os-user")
				},
			},
		},
		{
			name: "load error",
			fields: fields{
				Account: "012345678901",
				Region:  "us-east-1",
				OSUser:  "os-user",
			},
			args: args{
				id:  "11111111aaaaaaaa",
				key: []byte{48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 48, 49, 50, 51, 52, 53},
			},
			want: want{
				isError: true,
			},
			hook: hook{
				before: func() {
					keyring.MockInitWithError(testutil.NewError())
				},
				after: nil,
			},
		},
		{
			name: "conflict error",
			fields: fields{
				Account: "012345678901",
				Region:  "us-east-1",
				OSUser:  "os-user",
			},
			args: args{
				id:  "11111111aaaaaaaa",
				key: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			},
			want: want{
				isError: true,
			},
			hook: hook{
				before: func() {
					keyring.MockInit()
					buf := &bytes.Buffer{}
					_ = gob.NewEncoder(buf).Encode(NewMockState())
					_ = keyring.Set(params.CanonicalName, "os-user", buf.String())
				},
				after: func() {
					_ = keyring.Delete(params.CanonicalName, "os-user")
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.hook.before != nil {
				test.hook.before()
			}
			defer func() {
				if test.hook.after != nil {
					test.hook.after()
				}
			}()
			o := &Manager{
				Account: test.fields.Account,
				Region:  test.fields.Region,
				OSUser:  test.fields.OSUser,
				w:       test.fields.w,
			}
			err := o.AddCredential(test.args.id, test.args.key)
			testutil.CheckError(t, err != nil, test.want.isError)
		})
	}
}

func TestRemoveCredential(t *testing.T) {
	type fields struct {
		Account string
		Region  string
		OSUser  string
		w       io.Writer
	}
	type args struct {
		id  string
		key []byte
	}
	type want struct {
		isError bool
	}
	type hook struct {
		before func()
		after  func()
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
		hook   hook
	}{
		{
			name: "success",
			fields: fields{
				Account: "012345678901",
				Region:  "us-east-1",
				OSUser:  "os-user",
			},
			args: args{
				id:  "22222222bbbbbbbb",
				key: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			},
			want: want{
				isError: false,
			},
			hook: hook{
				before: func() {
					keyring.MockInit()
					buf := &bytes.Buffer{}
					_ = gob.NewEncoder(buf).Encode(NewMockState(func(s *State) {
						s.MasterKeys["22222222bbbbbbbb"] = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
					}))
					_ = keyring.Set(params.CanonicalName, "os-user", buf.String())
				},
				after: func() {
					_ = keyring.Delete(params.CanonicalName, "os-user")
				},
			},
		},
		{
			name: "load error",
			fields: fields{
				Account: "012345678901",
				Region:  "us-east-1",
				OSUser:  "os-user",
			},
			args: args{
				id:  "22222222bbbbbbbb",
				key: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			},
			want: want{
				isError: true,
			},
			hook: hook{
				before: func() {
					keyring.MockInitWithError(testutil.NewError())
				},
				after: nil,
			},
		},
		{
			name: "keyid error",
			fields: fields{
				Account: "012345678901",
				Region:  "us-east-1",
				OSUser:  "os-user",
			},
			args: args{
				id:  "11111111aaaaaaaa",
				key: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			},
			want: want{
				isError: true,
			},
			hook: hook{
				before: func() {
					keyring.MockInit()
					buf := &bytes.Buffer{}
					_ = gob.NewEncoder(buf).Encode(NewMockState())
					_ = keyring.Set(params.CanonicalName, "os-user", buf.String())
				},
				after: func() {
					_ = keyring.Delete(params.CanonicalName, "os-user")
				},
			},
		},
		{
			name: "conflict error",
			fields: fields{
				Account: "012345678901",
				Region:  "us-east-1",
				OSUser:  "os-user",
			},
			args: args{
				id:  "22222222bbbbbbbb",
				key: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 17},
			},
			want: want{
				isError: true,
			},
			hook: hook{
				before: func() {
					keyring.MockInit()
					buf := &bytes.Buffer{}
					_ = gob.NewEncoder(buf).Encode(NewMockState(func(s *State) {
						s.KeyID = "11111111aaaaaaaa"
						s.MasterKeys["22222222bbbbbbbb"] = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
					}))
					_ = keyring.Set(params.CanonicalName, "os-user", buf.String())
				},
				after: func() {
					_ = keyring.Delete(params.CanonicalName, "os-user")
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.hook.before != nil {
				test.hook.before()
			}
			defer func() {
				if test.hook.after != nil {
					test.hook.after()
				}
			}()
			o := &Manager{
				Account: test.fields.Account,
				Region:  test.fields.Region,
				OSUser:  test.fields.OSUser,
				w:       test.fields.w,
			}
			err := o.RemoveCredential(test.args.id, test.args.key)
			testutil.CheckError(t, err != nil, test.want.isError)
		})
	}
}

func TestListCredentials(t *testing.T) {
	type fields struct {
		Account string
		Region  string
		OSUser  string
		w       io.Writer
	}
	type want struct {
		value   []KeyEntry
		isError bool
	}
	type hook struct {
		before func()
		after  func()
	}
	tests := []struct {
		name   string
		fields fields
		want   want
		hook   hook
	}{
		{
			name: "success",
			fields: fields{
				Account: "012345678901",
				Region:  "us-east-1",
				OSUser:  "os-user",
			},
			want: want{
				value: []KeyEntry{
					{
						KeyID:    "11111111aaaaaaaa",
						IsActive: true,
					},
					{
						KeyID:    "22222222bbbbbbbb",
						IsActive: false,
					},
				},
				isError: false,
			},
			hook: hook{
				before: func() {
					keyring.MockInit()
					buf := &bytes.Buffer{}
					_ = gob.NewEncoder(buf).Encode(NewMockState(func(s *State) {
						s.MasterKeys["22222222bbbbbbbb"] = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
					}))
					_ = keyring.Set(params.CanonicalName, "os-user", buf.String())
				},
				after: func() {
					_ = keyring.Delete(params.CanonicalName, "os-user")
				},
			},
		},
		{
			name: "load error",
			fields: fields{
				Account: "012345678901",
				Region:  "us-east-1",
				OSUser:  "os-user",
			},
			want: want{
				value:   nil,
				isError: true,
			},
			hook: hook{
				before: func() {
					keyring.MockInitWithError(testutil.NewError())
				},
				after: nil,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.hook.before != nil {
				test.hook.before()
			}
			defer func() {
				if test.hook.after != nil {
					test.hook.after()
				}
			}()
			o := &Manager{
				Account: test.fields.Account,
				Region:  test.fields.Region,
				OSUser:  test.fields.OSUser,
				w:       test.fields.w,
			}
			got, err := o.ListCredentials()
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}

func TestManager_EnsureState(t *testing.T) {
	type fields struct {
		Account string
		Region  string
		OSUser  string
		sts     ISTS
		w       io.Writer
	}
	type args struct {
		cred string
	}
	type want struct {
		value   *State
		isError bool
	}
	type hook struct {
		before func()
		after  func()
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
		hook   hook
	}{
		{
			name: "success with credential",
			fields: fields{
				Account: "012345678901",
				Region:  "us-east-1",
				OSUser:  "os-user",
			},
			args: args{
				cred: "11111111aaaaaaaa30313233343536373839303132333435",
			},
			want: want{
				value:   NewMockState(),
				isError: false,
			},
			hook: hook{
				before: func() {
					keyring.MockInit()
				},
				after: func() {
					_ = keyring.Delete(params.CanonicalName, "os-user")
				},
			},
		},
		{
			name: "success without credential",
			fields: fields{
				Account: "012345678901",
				Region:  "us-east-1",
				OSUser:  "os-user",
			},
			args: args{
				cred: "",
			},
			want: want{
				value:   NewMockState(),
				isError: false,
			},
			hook: hook{
				before: func() {
					keyring.MockInit()
					buf := &bytes.Buffer{}
					_ = gob.NewEncoder(buf).Encode(NewMockState())
					_ = keyring.Set(params.CanonicalName, "os-user", buf.String())
				},
				after: func() {
					_ = keyring.Delete(params.CanonicalName, "os-user")
				},
			},
		},
		{
			name: "invalid credential",
			fields: fields{
				Account: "012345678901",
				Region:  "us-east-1",
				OSUser:  "os-user",
			},
			args: args{
				cred: "11111111aaaaaaaa30313233343536373839303132333435!",
			},
			want: want{
				value:   nil,
				isError: true,
			},
			hook: hook{
				before: func() {
					keyring.MockInit()
				},
				after: func() {
					_ = keyring.Delete(params.CanonicalName, "os-user")
				},
			},
		},
		{
			name: "error",
			fields: fields{
				Account: "012345678901",
				Region:  "us-east-1",
				OSUser:  "os-user",
			},
			args: args{
				cred: "",
			},
			want: want{
				value:   nil,
				isError: true,
			},
			hook: hook{
				before: func() {
					keyring.MockInitWithError(testutil.NewError())
					_ = keyring.Set(params.CanonicalName, "os-user", "data")
				},
				after: func() {
					_ = keyring.Delete(params.CanonicalName, "os-user")
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.hook.before != nil {
				test.hook.before()
			}
			defer func() {
				if test.hook.after != nil {
					test.hook.after()
				}
			}()
			o := &Manager{
				Account: test.fields.Account,
				Region:  test.fields.Region,
				OSUser:  test.fields.OSUser,
				sts:     test.fields.sts,
				w:       test.fields.w,
			}
			got, err := o.EnsureState(test.args.cred)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}

func TestManager_StoreState(t *testing.T) {
	type fields struct {
		Account string
		Region  string
		OSUser  string
		sts     ISTS
		w       io.Writer
	}
	type args struct {
		state *State
	}
	type want struct {
		isError bool
	}
	type hook struct {
		before func()
		after  func()
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
		hook   hook
	}{
		{
			name: "success",
			fields: fields{
				OSUser: "os-user",
			},
			args: args{
				state: NewMockState(),
			},
			want: want{
				isError: false,
			},
			hook: hook{
				before: func() {
					keyring.MockInit()
				},
			},
		},
		{
			name: "error",
			fields: fields{
				OSUser: "os-user",
			},
			args: args{
				state: NewMockState(),
			},
			want: want{
				isError: true,
			},
			hook: hook{
				before: func() {
					keyring.MockInitWithError(testutil.NewError())
					_ = keyring.Set(params.CanonicalName, "os-user", "data")
				},
				after: func() {
					_ = keyring.Delete(params.CanonicalName, "os-user")
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.hook.before != nil {
				test.hook.before()
			}
			defer func() {
				if test.hook.after != nil {
					test.hook.after()
				}
			}()
			o := &Manager{
				Account: test.fields.Account,
				Region:  test.fields.Region,
				OSUser:  test.fields.OSUser,
				sts:     test.fields.sts,
				w:       test.fields.w,
			}
			err := o.StoreState(test.args.state)
			testutil.CheckError(t, err != nil, test.want.isError)
		})
	}
}

func TestManager_LoadState(t *testing.T) {
	type fields struct {
		Account string
		Region  string
		OSUser  string
		sts     ISTS
		w       io.Writer
	}
	type want struct {
		value   *State
		isError bool
	}
	type hook struct {
		before func()
		after  func()
	}
	tests := []struct {
		name   string
		fields fields
		want   want
		hook   hook
	}{
		{
			name: "success",
			fields: fields{
				Account: "012345678901",
				Region:  "us-east-1",
				OSUser:  "os-user",
			},
			want: want{
				value:   NewMockState(),
				isError: false,
			},
			hook: hook{
				before: func() {
					keyring.MockInit()
					buf := &bytes.Buffer{}
					_ = gob.NewEncoder(buf).Encode(NewMockState())
					_ = keyring.Set(params.CanonicalName, "os-user", buf.String())
				},
				after: func() {
					_ = keyring.Delete(params.CanonicalName, "os-user")
				},
			},
		},
		{
			name: "load error",
			fields: fields{
				Account: "012345678901",
				Region:  "us-east-1",
				OSUser:  "os-user",
			},
			want: want{
				value:   nil,
				isError: true,
			},
			hook: hook{
				before: func() {
					keyring.MockInitWithError(testutil.NewError())
				},
				after: nil,
			},
		},
		{
			name: "verify account error",
			fields: fields{
				Account: "123456789012",
				Region:  "us-east-1",
				OSUser:  "os-user",
			},
			want: want{
				value:   nil,
				isError: true,
			},
			hook: hook{
				before: func() {
					keyring.MockInit()
					buf := &bytes.Buffer{}
					_ = gob.NewEncoder(buf).Encode(NewMockState())
					_ = keyring.Set(params.CanonicalName, "os-user", buf.String())
				},
				after: func() {
					_ = keyring.Delete(params.CanonicalName, "os-user")
				},
			},
		},
		{
			name: "verify region error",
			fields: fields{
				Account: "012345678901",
				Region:  "us-east-2",
				OSUser:  "os-user",
			},
			want: want{
				value:   nil,
				isError: true,
			},
			hook: hook{
				before: func() {
					keyring.MockInit()
					buf := &bytes.Buffer{}
					_ = gob.NewEncoder(buf).Encode(NewMockState())
					_ = keyring.Set(params.CanonicalName, "os-user", buf.String())
				},
				after: func() {
					_ = keyring.Delete(params.CanonicalName, "os-user")
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.hook.before != nil {
				test.hook.before()
			}
			defer func() {
				if test.hook.after != nil {
					test.hook.after()
				}
			}()
			o := &Manager{
				Account: test.fields.Account,
				Region:  test.fields.Region,
				OSUser:  test.fields.OSUser,
				sts:     test.fields.sts,
				w:       test.fields.w,
			}
			got, err := o.LoadState()
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}

func TestManager_CheckStateExist(t *testing.T) {
	type fields struct {
		Account string
		Region  string
		OSUser  string
		sts     ISTS
		w       io.Writer
	}
	type want struct {
		value   bool
		isError bool
	}
	type hook struct {
		before func()
		after  func()
	}
	tests := []struct {
		name   string
		fields fields
		want   want
		hook   hook
	}{
		{
			name: "success",
			fields: fields{
				Account: "012345678901",
				Region:  "us-east-1",
				OSUser:  "os-user",
			},
			want: want{
				value:   true,
				isError: false,
			},
			hook: hook{
				before: func() {
					keyring.MockInit()
					buf := &bytes.Buffer{}
					_ = gob.NewEncoder(buf).Encode(NewMockState())
					_ = keyring.Set(params.CanonicalName, "os-user", buf.String())
				},
				after: func() {
					_ = keyring.Delete(params.CanonicalName, "os-user")
				},
			},
		},
		{
			name: "load error",
			fields: fields{
				Account: "012345678901",
				Region:  "us-east-1",
				OSUser:  "os-user",
			},
			want: want{
				value:   false,
				isError: true,
			},
			hook: hook{
				before: func() {
					keyring.MockInitWithError(testutil.NewError())
				},
				after: nil,
			},
		},
		{
			name: "not found",
			fields: fields{
				Account: "012345678901",
				Region:  "us-east-1",
				OSUser:  "os-user",
			},
			want: want{
				value:   false,
				isError: false,
			},
			hook: hook{
				before: func() {
					keyring.MockInitWithError(keyring.ErrNotFound)
				},
				after: nil,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.hook.before != nil {
				test.hook.before()
			}
			defer func() {
				if test.hook.after != nil {
					test.hook.after()
				}
			}()
			o := &Manager{
				Account: test.fields.Account,
				Region:  test.fields.Region,
				OSUser:  test.fields.OSUser,
				sts:     test.fields.sts,
				w:       test.fields.w,
			}
			got, err := o.CheckStateExist()
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}

func TestManager_DeleteState(t *testing.T) {
	type fields struct {
		Account string
		Region  string
		OSUser  string
		sts     ISTS
		w       io.Writer
	}
	type want struct {
		isError bool
	}
	type hook struct {
		before func()
		after  func()
	}
	tests := []struct {
		name   string
		fields fields
		want   want
		hook   hook
	}{
		{
			name: "success",
			fields: fields{
				OSUser: "os-user",
			},
			want: want{
				isError: false,
			},
			hook: hook{
				before: func() {
					keyring.MockInit()
					_ = keyring.Set(params.CanonicalName, "os-user", "data")
				},
				after: nil,
			},
		},
		{
			name: "state does not exist",
			fields: fields{
				OSUser: "os-user",
			},
			want: want{
				isError: true,
			},
			hook: hook{
				before: func() {
					keyring.MockInit()
				},
				after: nil,
			},
		},
		{
			name: "other error",
			fields: fields{
				OSUser: "os-user",
			},
			want: want{
				isError: true,
			},
			hook: hook{
				before: func() {
					keyring.MockInitWithError(testutil.NewError())
				},
				after: nil,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.hook.before != nil {
				test.hook.before()
			}
			defer func() {
				if test.hook.after != nil {
					test.hook.after()
				}
			}()
			o := &Manager{
				Account: test.fields.Account,
				Region:  test.fields.Region,
				OSUser:  test.fields.OSUser,
				sts:     test.fields.sts,
				w:       test.fields.w,
			}
			err := o.DeleteState()
			testutil.CheckError(t, err != nil, test.want.isError)
		})
	}
}

func TestManager_GenerateState(t *testing.T) {
	type fields struct {
		Account string
		Region  string
		OSUser  string
		sts     ISTS
		w       io.Writer
	}
	type args struct {
		id  string
		key []byte
	}
	type want struct {
		value   *State
		isError bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "success",
			fields: fields{
				Account: "012345678901",
				Region:  "us-east-1",
				OSUser:  "os-user",
				sts:     nil,
				w:       nil,
			},
			args: args{
				id:  "11111111aaaaaaaa",
				key: []byte{48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 48, 49, 50, 51, 52, 53},
			},
			want: want{
				value:   NewMockState(),
				isError: false,
			},
		},
		{
			name: "no id provided",
			fields: fields{
				Account: "012345678901",
				Region:  "us-east-1",
				OSUser:  "os-user",
				sts:     nil,
				w:       nil,
			},
			args: args{
				id:  "",
				key: []byte{48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 48, 49, 50, 51, 52, 53},
			},
			want: want{
				value:   nil,
				isError: true,
			},
		},
		{
			name: "no key provided",
			fields: fields{
				Account: "012345678901",
				Region:  "us-east-1",
				OSUser:  "os-user",
				sts:     nil,
				w:       nil,
			},
			args: args{
				id:  "11111111aaaaaaaa",
				key: nil,
			},
			want: want{
				value:   nil,
				isError: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			o := &Manager{
				Account: test.fields.Account,
				Region:  test.fields.Region,
				OSUser:  test.fields.OSUser,
				sts:     test.fields.sts,
				w:       test.fields.w,
			}
			got, err := o.GenerateState(test.args.id, test.args.key)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}
