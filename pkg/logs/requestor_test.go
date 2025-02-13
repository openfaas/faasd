package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/openfaas/faas-provider/logs"
)

func Test_parseEntry(t *testing.T) {
	rawEntry := `{ "__CURSOR" : "s=71c4550142d14ace8e2959e3540cc15c;i=133c;b=44864010f0d94baba7b6bf8019f82a56;m=2945cd3;t=5a00d4eb59180;x=8ed47f7f9b3d798", "__REALTIME_TIMESTAMP" : "1583353899094400", "__MONOTONIC_TIMESTAMP" : "43277523", "_BOOT_ID" : "44864010f0d94baba7b6bf8019f82a56", "SYSLOG_IDENTIFIER" : "openfaas-fn:nodeinfo", "_PID" : "2254", "MESSAGE" : "2020/03/04 20:31:39 POST / - 200 OK - ContentLength: 83", "_SOURCE_REALTIME_TIMESTAMP" : "1583353899094372" }`
	expectedEntry := logs.Message{
		Name:      "nodeinfo",
		Namespace: "openfaas-fn",
		Text:      "2020/03/04 20:31:39 POST / - 200 OK - ContentLength: 83",
		Timestamp: time.Unix(0, 1583353899094400*1000).UTC(),
	}

	value := map[string]string{}
	json.Unmarshal([]byte(rawEntry), &value)

	entry, err := parseEntry(value)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	if entry.Name != expectedEntry.Name {
		t.Fatalf("want Name: %q, got %q", expectedEntry.Name, entry.Name)
	}

	if entry.Namespace != expectedEntry.Namespace {
		t.Fatalf("want Namespace: %q, got %q", expectedEntry.Namespace, entry.Namespace)
	}

	if entry.Timestamp != expectedEntry.Timestamp {
		t.Fatalf("want Timestamp: %q, got %q", expectedEntry.Timestamp, entry.Timestamp)
	}

	if entry.Text != expectedEntry.Text {
		t.Fatalf("want Text: %q, got %q", expectedEntry.Text, entry.Text)
	}
}

func Test_buildCmd(t *testing.T) {
	ctx := context.TODO()
	now := time.Now()
	req := logs.Request{
		Name:      "loggyfunc",
		Namespace: "spacetwo",
		Follow:    true,
		Since:     &now,
		Tail:      5,
	}

	expectedArgs := fmt.Sprintf(
		"--utc --no-pager --output=json --identifier=spacetwo:loggyfunc --since=%s --follow --lines=5",
		now.UTC().Format("2006-01-02 15:04:05"),
	)

	cmd := buildCmd(ctx, req).String()
	wantCmd := "journalctl"
	if !strings.Contains(cmd, wantCmd) {
		t.Fatalf("cmd want: %q, got: %q", wantCmd, cmd)
	}

	if !strings.HasSuffix(cmd, expectedArgs) {
		t.Fatalf("arg want: %q\ngot: %q", expectedArgs, cmd)
	}
}

// Test generated using Keploy
func TestQuery_JournalctlNotFound_Error(t *testing.T) {
	ctx := context.TODO()
	req := logs.Request{
		Name:      "testfunc",
		Namespace: "testnamespace",
	}

	// Temporarily override PATH to simulate `journalctl` not being found
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", originalPath)

	r := New()
	_, err := r.Query(ctx, req)
	if err == nil || !strings.Contains(err.Error(), "can not find journalctl") {
		t.Fatalf("expected error about missing journalctl, got: %v", err)
	}
}

// Test generated using Keploy
func TestParseEntry_InvalidSyslogIdentifier(t *testing.T) {
	rawEntry := map[string]string{
		"MESSAGE":              "Test message",
		"_PID":                 "1234",
		"SYSLOG_IDENTIFIER":    "invalididentifier",
		"__REALTIME_TIMESTAMP": "1583353899094400",
	}

	_, err := parseEntry(rawEntry)
	if err == nil || !strings.Contains(err.Error(), "invalid SYSLOG_IDENTIFIER") {
		t.Fatalf("expected error about invalid SYSLOG_IDENTIFIER, got: %v", err)
	}
}

// Test generated using Keploy
func TestParseEntry_MissingTimestamp(t *testing.T) {
	rawEntry := map[string]string{
		"MESSAGE":           "Test message",
		"_PID":              "1234",
		"SYSLOG_IDENTIFIER": "namespace:name",
	}

	_, err := parseEntry(rawEntry)
	if err == nil || !strings.Contains(err.Error(), "missing required field __REALTIME_TIMESTAMP") {
		t.Fatalf("expected error about missing __REALTIME_TIMESTAMP, got: %v", err)
	}
}
