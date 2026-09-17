package anyproxy_test

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCLIDoesNotLogProxyCredentials(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "anyproxy.exe")
	if out, err := exec.Command("go", "build", "-o", bin, "./cmd/anyproxy").CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}

	ssUserinfo := base64.StdEncoding.EncodeToString([]byte("aes-256-gcm:ss-secret"))
	secrets := []string{"p%40ss-secret", "p@ss-secret", ssUserinfo}

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "-a", "", "socks5://alice:p%40ss-secret@127.0.0.1:0", "ss://"+ssUserinfo+"@127.0.0.1:0")
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	var output strings.Builder
	var listenLine string
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintln(&output, line)
		if listenLine == "" && strings.Contains(line, " listen ") {
			listenLine = line
			cancel() // kills the server, which ends the pipe
		}
	}
	cmd.Wait()
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}

	if listenLine == "" {
		t.Fatalf("no listen line in output:\n%s", output.String())
	}
	if !strings.Contains(listenLine, "127.0.0.1:0") {
		t.Errorf("listen line lacks the listener address: %q", listenLine)
	}
	for _, secret := range secrets {
		if strings.Contains(output.String(), secret) {
			t.Errorf("output contains credential %q:\n%s", secret, output.String())
		}
	}
}
