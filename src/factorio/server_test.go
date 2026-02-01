package factorio_test

import (
	"testing"

	"github.com/OpenFactorioServerManager/factorio-server-manager/factorio"
)

func TestVersionString(t *testing.T) {
	tests := []struct {
		name    string
		version factorio.Version
		want    string
	}{
		{
			name:    "zero version",
			version: factorio.Version{0, 0, 0, 0},
			want:    "0.0.0.0",
		},
		{
			name:    "standard version",
			version: factorio.Version{1, 1, 0, 0},
			want:    "1.1.0.0",
		},
		{
			name:    "full version with build",
			version: factorio.Version{1, 1, 72, 123},
			want:    "1.1.72.123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.version.String()
			if got != tt.want {
				t.Errorf("Version.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVersionUnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		want    factorio.Version
		wantErr bool
	}{
		{
			name:    "two part version",
			text:    "1.1",
			want:    factorio.Version{1, 1, 0, 0},
			wantErr: false,
		},
		{
			name:    "three part version",
			text:    "1.1.72",
			want:    factorio.Version{1, 1, 72, 0},
			wantErr: false,
		},
		{
			name:    "full version",
			text:    "1.1.72.123",
			want:    factorio.Version{1, 1, 72, 123},
			wantErr: false,
		},
		{
			name:    "invalid version",
			text:    "abc.def",
			want:    factorio.Version{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v factorio.Version
			err := v.UnmarshalText([]byte(tt.text))
			if (err != nil) != tt.wantErr {
				t.Errorf("Version.UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !v.Equals(tt.want) {
				t.Errorf("Version.UnmarshalText() = %v, want %v", v, tt.want)
			}
		})
	}
}

func TestVersionEquals(t *testing.T) {
	tests := []struct {
		name string
		v1   factorio.Version
		v2   factorio.Version
		want bool
	}{
		{
			name: "equal versions",
			v1:   factorio.Version{1, 1, 0, 0},
			v2:   factorio.Version{1, 1, 0, 0},
			want: true,
		},
		{
			name: "different major",
			v1:   factorio.Version{1, 1, 0, 0},
			v2:   factorio.Version{2, 1, 0, 0},
			want: false,
		},
		{
			name: "different minor",
			v1:   factorio.Version{1, 1, 0, 0},
			v2:   factorio.Version{1, 2, 0, 0},
			want: false,
		},
		{
			name: "different patch",
			v1:   factorio.Version{1, 1, 0, 0},
			v2:   factorio.Version{1, 1, 1, 0},
			want: false,
		},
		{
			name: "different build",
			v1:   factorio.Version{1, 1, 0, 0},
			v2:   factorio.Version{1, 1, 0, 1},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.v1.Equals(tt.v2)
			if got != tt.want {
				t.Errorf("Version.Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionLess(t *testing.T) {
	tests := []struct {
		name string
		v1   factorio.Version
		v2   factorio.Version
		want bool
	}{
		{
			name: "equal versions not less",
			v1:   factorio.Version{1, 1, 0, 0},
			v2:   factorio.Version{1, 1, 0, 0},
			want: false,
		},
		{
			name: "major less",
			v1:   factorio.Version{1, 1, 0, 0},
			v2:   factorio.Version{2, 0, 0, 0},
			want: true,
		},
		{
			name: "major greater",
			v1:   factorio.Version{2, 0, 0, 0},
			v2:   factorio.Version{1, 1, 0, 0},
			want: false,
		},
		{
			name: "minor less",
			v1:   factorio.Version{1, 0, 0, 0},
			v2:   factorio.Version{1, 1, 0, 0},
			want: true,
		},
		{
			name: "patch less",
			v1:   factorio.Version{1, 1, 0, 0},
			v2:   factorio.Version{1, 1, 1, 0},
			want: true,
		},
		{
			name: "build less",
			v1:   factorio.Version{1, 1, 0, 0},
			v2:   factorio.Version{1, 1, 0, 1},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.v1.Less(tt.v2)
			if got != tt.want {
				t.Errorf("Version.Less() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionGreater(t *testing.T) {
	tests := []struct {
		name string
		v1   factorio.Version
		v2   factorio.Version
		want bool
	}{
		{
			name: "equal versions not greater",
			v1:   factorio.Version{1, 1, 0, 0},
			v2:   factorio.Version{1, 1, 0, 0},
			want: false,
		},
		{
			name: "major greater",
			v1:   factorio.Version{2, 0, 0, 0},
			v2:   factorio.Version{1, 1, 0, 0},
			want: true,
		},
		{
			name: "major less not greater",
			v1:   factorio.Version{1, 1, 0, 0},
			v2:   factorio.Version{2, 0, 0, 0},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.v1.Greater(tt.v2)
			if got != tt.want {
				t.Errorf("Version.Greater() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionCompatible(t *testing.T) {
	tests := []struct {
		name string
		v1   factorio.Version
		v2   factorio.Version
		op   string
		want bool
	}{
		{
			name: "equal with ==",
			v1:   factorio.Version{1, 1, 0, 0},
			v2:   factorio.Version{1, 1, 0, 0},
			op:   "==",
			want: true,
		},
		{
			name: "not equal with ==",
			v1:   factorio.Version{1, 1, 0, 0},
			v2:   factorio.Version{1, 2, 0, 0},
			op:   "==",
			want: false,
		},
		{
			name: "not equal with !=",
			v1:   factorio.Version{1, 1, 0, 0},
			v2:   factorio.Version{1, 2, 0, 0},
			op:   "!=",
			want: true,
		},
		{
			name: "greater with >",
			v1:   factorio.Version{2, 0, 0, 0},
			v2:   factorio.Version{1, 0, 0, 0},
			op:   ">",
			want: true,
		},
		{
			name: "less with <",
			v1:   factorio.Version{1, 0, 0, 0},
			v2:   factorio.Version{2, 0, 0, 0},
			op:   "<",
			want: true,
		},
		{
			name: "greater or equal with >= (equal)",
			v1:   factorio.Version{1, 1, 0, 0},
			v2:   factorio.Version{1, 1, 0, 0},
			op:   ">=",
			want: true,
		},
		{
			name: "greater or equal with >= (greater)",
			v1:   factorio.Version{2, 0, 0, 0},
			v2:   factorio.Version{1, 0, 0, 0},
			op:   ">=",
			want: true,
		},
		{
			name: "less or equal with <= (equal)",
			v1:   factorio.Version{1, 1, 0, 0},
			v2:   factorio.Version{1, 1, 0, 0},
			op:   "<=",
			want: true,
		},
		{
			name: "unsupported operator returns false",
			v1:   factorio.Version{1, 1, 0, 0},
			v2:   factorio.Version{1, 1, 0, 0},
			op:   "???",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.v1.Compatible(tt.v2, tt.op)
			if got != tt.want {
				t.Errorf("Version.Compatible(%v, %q) = %v, want %v", tt.v2, tt.op, got, tt.want)
			}
		})
	}
}

func TestNilVersion(t *testing.T) {
	nilVersion := factorio.NilVersion
	expected := factorio.Version{0, 0, 0, 0}

	if !nilVersion.Equals(expected) {
		t.Errorf("NilVersion = %v, want %v", nilVersion, expected)
	}
}

func TestServerRunningState(t *testing.T) {
	t.Skip("Server tests require Factorio binary - skipping in unit tests")

	// These tests would test:
	// - SetRunning and GetRunning
	// - Server lifecycle
	// - WebSocket notifications on state change
}
