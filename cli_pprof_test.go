package anyproxy_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCLIPprof(t *testing.T) {
	exe := filepath.Join(t.TempDir(), "anyproxy.exe")
	build := exec.CommandContext(t.Context(), "go", "build", "-o", exe, "./cmd/anyproxy")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}

	t.Run("default listeners do not serve pprof", func(t *testing.T) {
		addr := freeAddr(t)
		startCLI(t, exe, addr, "-a", addr)
		status, body := get(t, "http://"+addr+"/debug/pprof/cmdline", nil)
		if status == http.StatusOK || strings.Contains(body, exe) {
			t.Errorf("pprof exposed: status %d body %q", status, body)
		}
		origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, "origin")
		}))
		defer origin.Close()
		status, body = get(t, origin.URL, &url.URL{Scheme: "http", Host: addr})
		if status != http.StatusOK || body != "origin" {
			t.Errorf("proxied GET: status %d body %q", status, body)
		}
	})

	t.Run("explicit pprof listener serves cmdline", func(t *testing.T) {
		addr := freeAddr(t)
		startCLI(t, exe, addr, "-a", "", "pprof://"+addr)
		status, body := get(t, "http://"+addr+"/debug/pprof/cmdline", nil)
		want := strings.Join([]string{exe, "-a", "", "pprof://" + addr}, "\x00")
		if status != http.StatusOK || body != want {
			t.Errorf("status %d body %q, want 200 %q", status, body, want)
		}
	})
}

func freeAddr(t *testing.T) string {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	listener.Close()
	return listener.Addr().String()
}

func startCLI(t *testing.T, exe, addr string, args ...string) {
	ctx, cancel := context.WithCancel(t.Context())
	cmd := exec.CommandContext(ctx, exe, args...)
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cancel()
		cmd.Wait()
	})
	dialCtx, dialCancel := context.WithTimeout(ctx, 5*time.Second)
	defer dialCancel()
	var dialer net.Dialer
	for {
		conn, err := dialer.DialContext(dialCtx, "tcp", addr)
		if err == nil {
			conn.Close()
			return
		}
		select {
		case <-dialCtx.Done():
			t.Fatalf("%s not listening: %v", addr, err)
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func get(t *testing.T, target string, proxy *url.URL) (int, string) {
	client := http.Client{
		Timeout:   5 * time.Second,
		Transport: &http.Transport{Proxy: http.ProxyURL(proxy)},
	}
	resp, err := client.Get(target)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, string(body)
}
