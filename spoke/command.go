package spoke

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/hekmon/transmissionrpc/v2"
)

type Spoke struct {
	command []string
	timeout time.Duration
}

func NewSpoke(timeout time.Duration, command []string) *Spoke {
	return &Spoke{
		command: command,
		timeout: timeout,
	}
}

func (s *Spoke) Query(ctx context.Context, t *transmissionrpc.Torrent) (string, error) {
	if len(s.command) == 0 {
		return "", errors.New("empty command provided to CommandNotifier")
	}

	if s.timeout != 0 {
		ctx1, cl := context.WithTimeout(ctx, s.timeout)
		defer cl()
		ctx = ctx1
	}

	subprocess := exec.CommandContext(ctx, s.command[0], s.command[1:]...)
	subprocess.Stderr = os.Stderr

	txPipe, err := subprocess.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("unable to get subprocess stdin pipe: %w", err)
	}
	defer txPipe.Close()

	go func() {
		defer txPipe.Close()
		enc := json.NewEncoder(txPipe)
		enc.SetIndent("", "\t")
		if err := enc.Encode(t); err != nil {
			log.Printf("spoke %#v: unable to export record: %v", s.command, err)
		}
	}()

	res, err := subprocess.Output()
	if err != nil {
		return "", fmt.Errorf("spoke %#v: subprocess invocation failed: %w", s.command, err)
	}
	return string(bytes.TrimRight(res, "\r\n")), nil
}
