package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTOrderStatus_Scan(t *testing.T) {
	tests := []struct {
		name    string
		status  TOrderStatus
		scanSrc interface{}
		wantErr bool
	}{
		{
			name:    "ok - NEW",
			status:  TOrderStatusNEW,
			scanSrc: "NEW",
			wantErr: false,
		},
		{
			name:    "ok - byte PROCESSING",
			status:  TOrderStatusPROCESSING,
			scanSrc: []byte("PROCESSING"),
			wantErr: false,
		},
		{
			name:    "fail - int",
			scanSrc: 1,
			wantErr: true,
		},
		{
			name:    "fail - bool",
			scanSrc: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var status TOrderStatus
			err := status.Scan(tt.scanSrc)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.status, status)
		})
	}
}

func TestNullTOrderStatus_Scan(t *testing.T) {
	tests := []struct {
		name    string
		status  TOrderStatus
		scanSrc interface{}
		wantErr bool
		valid   bool
	}{
		{
			name:    "ok - NEW",
			status:  TOrderStatusNEW,
			scanSrc: "NEW",
			wantErr: false,
			valid:   true,
		},
		{
			name:    "ok - byte PROCESSING",
			status:  TOrderStatusPROCESSING,
			scanSrc: []byte("PROCESSING"),
			wantErr: false,
			valid:   true,
		},
		{
			name:    "ok - null",
			scanSrc: nil,
			wantErr: false,
			valid:   false,
		},
		{
			name:    "fail - int",
			scanSrc: 1,
			wantErr: true,
			valid:   true,
		},
		{
			name:    "fail - bool",
			scanSrc: true,
			wantErr: true,
			valid:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var status NullTOrderStatus
			err := status.Scan(tt.scanSrc)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.status, status.TOrderStatus)
			assert.Equal(t, tt.valid, status.Valid)
		})
	}
}

func TestNullTOrderStatus_Value(t *testing.T) {
	tests := []struct {
		name    string
		status  NullTOrderStatus
		want    interface{}
		wantErr bool
	}{
		{
			name: "ok - valid status",
			status: NullTOrderStatus{
				TOrderStatus: TOrderStatusPROCESSING,
				Valid:        true,
			},
			want:    "PROCESSING",
			wantErr: false,
		},
		{
			name: "ok - null status",
			status: NullTOrderStatus{
				TOrderStatus: TOrderStatus(""),
				Valid:        false,
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.status.Value()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
