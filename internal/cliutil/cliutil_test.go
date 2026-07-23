package cliutil

import (
	"errors"
	"strings"
	"testing"
)

func TestReportErrorNonInteractiveDoesNotWait(t *testing.T) {
	var output strings.Builder
	reportError(errors.New("failure"), nil, &output, false)

	if got, want := output.String(), "エラー: failure\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestReportErrorInteractiveWaitsForInput(t *testing.T) {
	var output strings.Builder
	reportError(errors.New("failure"), strings.NewReader("\n"), &output, true)

	if got, want := output.String(), "エラー: failure\nEnterキーを押すと終了します...\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}
