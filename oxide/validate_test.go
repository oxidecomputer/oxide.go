// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidator_HasRequiredStr(t *testing.T) {
	type fields struct {
		err error
	}
	type args struct {
		value string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr error
	}{
		{
			name:   "string is present",
			fields: fields{},
			args: args{
				value: "some string",
			},
			want: true,
		},
		{
			name:   "string is not present",
			fields: fields{},
			args: args{
				value: "",
			},
			want:    false,
			wantErr: errors.Join(errors.New("required value is an empty string")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Validator{
				err: tt.fields.err,
			}
			got := v.HasRequiredStr(tt.args.value)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, v.err)

		})
	}
}

func TestValidator_HasRequiredObj(t *testing.T) {
	val := "hi"
	type fields struct {
		err error
	}
	type args struct {
		value any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr error
	}{
		{
			name:   "object is present",
			fields: fields{},
			args: args{
				value: &val,
			},
			want: true,
		},
		{
			name:   "object is not present",
			fields: fields{},
			args: args{
				value: nil,
			},
			want:    false,
			wantErr: errors.Join(errors.New("required value is nil")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Validator{
				err: tt.fields.err,
			}
			got := v.HasRequiredObj(tt.args.value)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, v.err)

		})
	}
}

func TestValidator_HasRequiredNum(t *testing.T) {
	type fields struct {
		err error
	}
	type args struct {
		value int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr error
	}{
		{
			name:   "int is present",
			fields: fields{},
			args: args{
				value: 1,
			},
			want: true,
		},
		{
			name:    "int is not present",
			fields:  fields{},
			args:    args{},
			want:    false,
			wantErr: errors.Join(errors.New("required value is zero")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Validator{
				err: tt.fields.err,
			}
			got := v.HasRequiredNum(tt.args.value)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, v.err)

		})
	}
}
