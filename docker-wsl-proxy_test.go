package main

import (
	"gotest.tools/assert"
	"testing"
)

func TestRewriteBindToWSL(t *testing.T) {
	testCases := []struct {
		input, expected string
	}{
		// See https://github.com/docker/for-win/issues/6628#issuecomment-635906394
		{`C:\windows\system32:/c`, `/mnt/c/windows/system32:/c`},
		{`C://windows/system32:/c`, `/mnt/c/windows/system32:/c`},
		{`//C/windows/system32:/c`, `/mnt/c/windows/system32:/c`},
		{`/C/windows/system32:/c`, `/mnt/c/windows/system32:/c`},
		{`/mnt/c/windows/system32:/c`, `/mnt/c/windows/system32:/c`},
		// Compatibility with Docker Desktop
		{`/host_mnt/c/windows/system32:/c`, `/mnt/c/windows/system32:/c`},
		// Additional cases
		{`C:\:/c`, `/mnt/c:/c`},
		{`C:/:/c`, `/mnt/c:/c`},
		{`/proc:/proc`, `/proc:/proc`},
		{`/:/host`, `/:/host`},
		// See https://github.com/slonopotamus/stevedore/issues/38#issuecomment-1076167200
		{`C:\Windows\System32\cmd.exe:/cmd.exe`, `/mnt/c/windows/system32/cmd.exe:/cmd.exe`},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			actual := rewriteBindToWSL(tc.input)
			assert.Equal(t, actual, tc.expected)
		})
	}
}

func TestRewriteBindToWindows(t *testing.T) {
	testCases := []struct {
		input, expected string
	}{
		{`/mnt/c/windows/system32:/c`, `C:\windows\system32:/c`},
		{`/proc:/proc`, `/proc:/proc`},
		{`/:/host`, `/:/host`},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			actual := rewriteBindToWindows(tc.input)
			assert.Equal(t, actual, tc.expected)
		})
	}
}
