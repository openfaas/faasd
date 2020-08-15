package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/openfaas/faas-provider/logs"

	faasd "github.com/openfaas/faasd/pkg"
)

type requester struct{}

// New returns a new journalctl log Requester
func New() logs.Requester {
	return &requester{}
}

// Query submits a log request to the actual logging system.
func (r *requester) Query(ctx context.Context, req logs.Request) (<-chan logs.Message, error) {
	ctx, cancel := context.WithCancel(ctx)
	_, err := exec.LookPath("journalctl")
	if err != nil {
		cancel()
		return nil, fmt.Errorf("can not find journalctl: %w", err)
	}

	cmd := buildCmd(ctx, req)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create journalctl pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create journalctl err pipe: %w", err)
	}

	err = cmd.Start()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create journalctl: %w", err)
	}

	// call start and get the stdout prior to streaming so that we can return a meaningful
	// error for as long as possible. If the cmd starts correctly, we are highly likely to
	// succeed anyway
	msgs := make(chan logs.Message)

	log.Println("starting journal stream using ", cmd.String())
	// will ensure `stdout` is closed and all related resources cleaned up
	go func() {
		err := cmd.Wait()
		log.Println("wait result", err)
		cancel()
	}()
	go streamLogs(ctx, stdout, msgs)
	go logErrOut(stderr)

	return msgs, nil
}

// buildCmd reeturns the equivalent of
//
// 	journalctl -t <namespace>:<name>  \
// 		--output=json \
// 		--since=<timestamp> \
// 		<--follow> \
func buildCmd(ctx context.Context, req logs.Request) *exec.Cmd {
	// // set the cursor position based on req, default to 5m
	since := time.Now().Add(-5 * time.Minute)
	if req.Since != nil && req.Since.Before(time.Now()) {
		since = *req.Since
	}

	namespace := req.Namespace
	if namespace == "" {
		namespace = faasd.FunctionNamespace
	}

	// find the description of the fields here
	// https://www.freedesktop.org/software/systemd/man/systemd.journal-fields.html
	// the available fields can vary greatly, the selected fields were detemined by
	// trial and error with journalctl in an ubuntu VM  (via multipass)
	args := []string{
		"--utc",
		"--no-pager",
		"--output=json",
		"--identifier=" + namespace + ":" + req.Name,
		fmt.Sprintf("--since=%s", since.UTC().Format("2006-01-02 15:04:05")),
	}

	if req.Follow {
		args = append(args, "--follow")
	}

	if req.Tail > 0 {
		args = append(args, fmt.Sprintf("--lines=%d", req.Tail))
	}

	return exec.CommandContext(ctx, "journalctl", args...)
}

// streamLogs copies log entries from the journalctl `cmd`/`out` to `msgs`
// the loop is based on the Decoder example in the docs
// https://golang.org/pkg/encoding/json/#Decoder.Decode
func streamLogs(ctx context.Context, out io.ReadCloser, msgs chan logs.Message) {
	log.Println("log stream started")

	defer func() {
		log.Println("closing journal stream")
		close(msgs)
	}()

	dec := json.NewDecoder(out)
	for ctx.Err() == nil {

		// wait until the output has filled with some json, this is required so that
		// we don't have any race condition with journalctl filling stdout
		// if we do not wait, `out` is empty and `dec.More()` is false and we will never
		// attempt to decode and then immediately close the stream.
		foundMore := false
		for !foundMore && ctx.Err() == nil {
			foundMore = dec.More()
			time.Sleep(time.Millisecond)
		}

		if ctx.Err() != nil {
			log.Println("log stream context cancelled")
			return
		}

		// the journalctl outputs all the values as a string, so a struct with json
		// tags wont help much
		entry := map[string]string{}
		err := dec.Decode(&entry)
		if err == io.EOF {
			// stream is done
			log.Println("log stream ended")
			return
		}
		if err != nil {
			log.Printf("error decoding journalctl output: %s", err)
			return
		}

		msg, err := parseEntry(entry)
		if err != nil {
			log.Printf("error parsing journalctl output: %s", err)
			return
		}

		msgs <- msg
	}
}

// parseEntry reads the deserialized json from journalctl into a log.Message
//
// The following fields are parsed from the journal
// - MESSAGE
// - _PID
// - SYSLOG_IDENTIFIER
// - __REALTIME_TIMESTAMP
func parseEntry(entry map[string]string) (logs.Message, error) {
	logMsg := logs.Message{
		Text:     entry["MESSAGE"],
		Instance: entry["_PID"],
	}

	identifier := entry["SYSLOG_IDENTIFIER"]
	parts := strings.Split(identifier, ":")
	if len(parts) != 2 {
		return logMsg, fmt.Errorf("invalid SYSLOG_IDENTIFIER")
	}
	logMsg.Namespace = parts[0]
	logMsg.Name = parts[1]

	ts, ok := entry["__REALTIME_TIMESTAMP"]
	if !ok {
		return logMsg, fmt.Errorf("missing required field __REALTIME_TIMESTAMP")
	}

	ms, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return logMsg, fmt.Errorf("invalid timestamp: %w", err)
	}
	logMsg.Timestamp = time.Unix(0, ms*1000).UTC()

	return logMsg, nil
}

func logErrOut(out io.ReadCloser) {
	defer log.Println("stderr closed")
	defer out.Close()

	_, _ = io.Copy(log.Writer(), out)
}
