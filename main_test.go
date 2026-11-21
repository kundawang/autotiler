/*
 * MIT License
 *
 * Copyright (c) 2024 The autotiler authors
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package main

import (
	"strings"
	"testing"
)

func TestSafeOutputPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"plain", "out/output.png", false},
		{"current_dir", "output.png", false},
		{"nested", "out/nested/dir/output.png", false},
		{"dot_cleaned", "./out/../out/output.png", false},
		{"escape_parent", "../output.png", true},
		{"escape_deep", "out/../../output.png", true},
		{"absolute_tmp", "/tmp/output.png", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := safeOutputPath(tt.path)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error for path %q", tt.path)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error for path %q: %v", tt.path, err)
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), "escapes the working directory") {
				t.Fatalf("error for path %q does not mention escaping: %v", tt.path, err)
			}
		})
	}
}

func TestExportMissingSourceImage(t *testing.T) {
	err := export(map[string][]string{}, "no/such/source.png", "out/should_not_exist.png")
	if err == nil {
		t.Fatal("expected error for missing source image")
	}
	if !strings.Contains(err.Error(), "no/such/source.png") {
		t.Fatalf("error does not name the missing source image: %v", err)
	}
}

func TestParseBool(t *testing.T) {
	for _, v := range []string{"1", "true", "on", "yes", "TRUE", "On"} {
		got, err := parseBool(v)
		if err != nil || !got {
			t.Errorf("parseBool(%q) = %v, %v; want true, nil", v, got, err)
		}
	}
	for _, v := range []string{"0", "false", "off", "no", "FALSE", "Off"} {
		got, err := parseBool(v)
		if err != nil || got {
			t.Errorf("parseBool(%q) = %v, %v; want false, nil", v, got, err)
		}
	}
	if _, err := parseBool("maybe"); err == nil {
		t.Error("expected error for invalid value")
	}
}
