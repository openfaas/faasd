package logs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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

func Test_streamLogs_withDelayedLogStreamOutput(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var b bytes.Buffer
	out := ioutil.NopCloser(&b)

	msgs := make(chan logs.Message)
	go streamLogs(ctx, out, msgs)

	time.Sleep(500 * time.Millisecond)
	n, err := b.Write([]byte(`{ "__CURSOR" : "s=9be00b9a7f6a4a0e93020e9ec5a688c4;i=d24;b=ee9ef3d8a1c24299935c0327e06714bf;m=118a4e86a;t=5acebc02e069f;x=b9a854940506d6ac", "__REALTIME_TIMESTAMP" : "1597503425087135", "__MONOTONIC_TIMESTAMP" : "4708427882", "_BOOT_ID" : "ee9ef3d8a1c24299935c0327e06714bf", "_MACHINE_ID" : "47f476b755cb4cd3878f20cb339718f2", "PRIORITY" : "3", "_UID" : "0", "_GID" : "0", "_SELINUX_CONTEXT" : "unconfined\n", "_SYSTEMD_SLICE" : "system.slice", "_TRANSPORT" : "journal", "_CAP_EFFECTIVE" : "3fffffffff", "_HOSTNAME" : "primary", "_SYSTEMD_CGROUP" : "/system.slice/containerd.service", "_SYSTEMD_UNIT" : "containerd.service", "_SYSTEMD_INVOCATION_ID" : "01528ec3a87242c382c3a6b2e5df7567", "_COMM" : "faasd", "_EXE" : "/usr/local/bin/faasd", "SYSLOG_IDENTIFIER" : "openfaas-fn:nodeinfo", "_CMDLINE" : "/usr/local/bin/faasd", "_PID" : "15278", "MESSAGE" : "2020/08/15 14:57:05 POST / - 200 OK - ContentLength: 85", "_SOURCE_REALTIME_TIMESTAMP" : "1597503425087112" }`))
	if err != nil {
		t.Fatalf("failed to write test log line: %s", err)
	}

	log.Printf("log line length %d\n", n)

	line := <-msgs
	if line.Name != "nodeinfo" {
		t.Fatalf("wrong function name; expected: %q, got:%q", "nodeinfo", line.Name)
	}

	if line.Text != "2020/08/15 14:57:05 POST / - 200 OK - ContentLength: 85" {
		t.Fatalf("wrong log text; expected: %q, got:%q", "2020/08/15 14:57:05 POST / - 200 OK - ContentLength: 85", line.Text)
	}

}
