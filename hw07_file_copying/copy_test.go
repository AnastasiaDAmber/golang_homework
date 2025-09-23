package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCopy(t *testing.T) {
	tests := []struct {
		name   string
		offset int64
		limit  int64
		golden string
	}{
		{"offset0_limit0", 0, 0, "testdata/out_offset0_limit0.txt"},
		{"offset0_limit10", 0, 10, "testdata/out_offset0_limit10.txt"},
		{"offset0_limit1000", 0, 1000, "testdata/out_offset0_limit1000.txt"},
		{"offset0_limit10000", 0, 10000, "testdata/out_offset0_limit10000.txt"},
		{"offset100_limit1000", 100, 1000, "testdata/out_offset100_limit1000.txt"},
		{"offset6000_limit1000", 6000, 1000, "testdata/out_offset6000_limit1000.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(os.TempDir(), "out.txt")
			defer os.Remove(tmpFile)

			err := Copy("testdata/input.txt", tmpFile, tt.offset, tt.limit)
			if err != nil {
				t.Fatalf("Copy failed: %v", err)
			}

			got, _ := os.ReadFile(tmpFile)
			want, _ := os.ReadFile(tt.golden)

			if string(got) != string(want) {
				t.Errorf("mismatch in %s\n got:  %q\n want: %q", tt.name, got, want)
			}
		})
	}
}

func TestErrors(t *testing.T) {
	// offset больше размера файла
	err := Copy("testdata/input.txt", filepath.Join(os.TempDir(), "out.txt"), 999999, 0)
	if !errors.Is(err, ErrOffsetExceedsFileSize) {
		t.Errorf("expected ErrOffsetExceedsFileSize, got %v", err)
	}

	// несуществующий файл
	err = Copy("no_such_file.txt", filepath.Join(os.TempDir(), "out.txt"), 0, 0)
	if err == nil {
		t.Errorf("expected error for missing file, got nil")
	}
}
