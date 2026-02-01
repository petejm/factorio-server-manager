package api_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/OpenFactorioServerManager/factorio-server-manager/api"
)

func TestValidatePathComponent(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:    "valid filename",
			input:   "save.zip",
			wantErr: nil,
		},
		{
			name:    "valid filename with underscores",
			input:   "my_save_file.zip",
			wantErr: nil,
		},
		{
			name:    "valid filename with hyphens",
			input:   "my-save-file.zip",
			wantErr: nil,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: api.ErrInvalidPath,
		},
		{
			name:    "double dot traversal",
			input:   "..",
			wantErr: api.ErrPathTraversal,
		},
		{
			name:    "path with double dot",
			input:   "../etc/passwd",
			wantErr: api.ErrPathTraversal,
		},
		{
			name:    "path with embedded double dot",
			input:   "foo/../bar",
			wantErr: api.ErrPathTraversal,
		},
		{
			name:    "forward slash",
			input:   "foo/bar",
			wantErr: api.ErrPathTraversal,
		},
		{
			name:    "backslash",
			input:   "foo\\bar",
			wantErr: api.ErrPathTraversal,
		},
		{
			name:    "single dot",
			input:   ".",
			wantErr: api.ErrPathTraversal,
		},
		{
			name:    "absolute path",
			input:   "/etc/passwd",
			wantErr: api.ErrPathTraversal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := api.ValidatePathComponent(tt.input)
			if err != tt.wantErr {
				t.Errorf("api.ValidatePathComponent(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestSafeJoinPath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "pathutil_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name      string
		baseDir   string
		userInput string
		wantErr   bool
	}{
		{
			name:      "valid join",
			baseDir:   tempDir,
			userInput: "save.zip",
			wantErr:   false,
		},
		{
			name:      "valid filename with extension",
			baseDir:   tempDir,
			userInput: "my_save.zip",
			wantErr:   false,
		},
		{
			name:      "traversal attempt with double dot",
			baseDir:   tempDir,
			userInput: "../etc/passwd",
			wantErr:   true,
		},
		{
			name:      "traversal attempt with slash",
			baseDir:   tempDir,
			userInput: "foo/bar",
			wantErr:   true,
		},
		{
			name:      "empty input",
			baseDir:   tempDir,
			userInput: "",
			wantErr:   true,
		},
		{
			name:      "just double dot",
			baseDir:   tempDir,
			userInput: "..",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := api.SafeJoinPath(tt.baseDir, tt.userInput)
			if (err != nil) != tt.wantErr {
				t.Errorf("api.SafeJoinPath(%q, %q) error = %v, wantErr %v", tt.baseDir, tt.userInput, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Verify the result is within the base directory
				absBase, _ := filepath.Abs(tt.baseDir)
				if result != filepath.Join(absBase, tt.userInput) {
					t.Errorf("api.SafeJoinPath(%q, %q) = %q, want %q", tt.baseDir, tt.userInput, result, filepath.Join(absBase, tt.userInput))
				}
			}
		})
	}
}

func TestValidateFileExtension(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		expectedExt string
		want        bool
	}{
		{
			name:        "valid zip extension",
			filename:    "save.zip",
			expectedExt: ".zip",
			want:        true,
		},
		{
			name:        "valid zip extension uppercase",
			filename:    "save.ZIP",
			expectedExt: ".zip",
			want:        true,
		},
		{
			name:        "valid zip mixed case",
			filename:    "save.ZiP",
			expectedExt: ".zip",
			want:        true,
		},
		{
			name:        "wrong extension",
			filename:    "save.tar",
			expectedExt: ".zip",
			want:        false,
		},
		{
			name:        "no extension",
			filename:    "save",
			expectedExt: ".zip",
			want:        false,
		},
		{
			name:        "double extension checks last",
			filename:    "save.tar.zip",
			expectedExt: ".zip",
			want:        true,
		},
		{
			name:        "hidden file with extension",
			filename:    ".hidden.zip",
			expectedExt: ".zip",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := api.ValidateFileExtension(tt.filename, tt.expectedExt)
			if got != tt.want {
				t.Errorf("api.ValidateFileExtension(%q, %q) = %v, want %v", tt.filename, tt.expectedExt, got, tt.want)
			}
		})
	}
}

func TestSecurityTraversalAttempts(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "security_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Common path traversal attack patterns
	attacks := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\config\\sam",
		"....//....//....//etc/passwd",
		"..%2f..%2f..%2fetc/passwd",
		"..%252f..%252f..%252fetc/passwd",
		"%2e%2e%2f%2e%2e%2f%2e%2e%2fetc/passwd",
		"..././..././..././etc/passwd",
		"..%c0%af..%c0%af..%c0%afetc/passwd",
		"..%c1%9c..%c1%9c..%c1%9cetc/passwd",
		"/etc/passwd",
		"\\etc\\passwd",
		"....//",
		"..../",
		"....\\",
	}

	for _, attack := range attacks {
		t.Run("attack_"+attack, func(t *testing.T) {
			_, err := api.SafeJoinPath(tempDir, attack)
			if err == nil {
				t.Errorf("api.SafeJoinPath should have rejected attack pattern: %q", attack)
			}
		})
	}
}
