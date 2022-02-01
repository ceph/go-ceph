package log_test

import (
	"io"
	stdlog "log"
	"os"
	"strings"
	"testing"

	"github.com/ceph/go-ceph/common/log"
	intLog "github.com/ceph/go-ceph/internal/log"
	"github.com/stretchr/testify/assert"
)

func testLog() {
	intLog.Debugf("-%s-", "debug")
	intLog.Warnf("-%s-", "warn")
}

var testOut = []string{
	"<go-ceph>[DBG]log_test.go:16: -debug-",
	"<go-ceph>[WRN]log_test.go:17: -warn-",
	"",
}

func checkLines(t *testing.T, lines []string) {
	for i := range lines {
		assert.Equal(t, testOut[len(testOut)-len(lines)+i], lines[i])
	}
}

func captureAllOutput(f func()) string {
	oldout := os.Stdout
	olderr := os.Stderr
	oldlog := stdlog.Writer()
	defer func() {
		os.Stdout = oldout
		os.Stderr = olderr
		stdlog.SetOutput(oldlog)
	}()
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w
	stdlog.SetOutput(w)
	go func() {
		f()
		_ = w.Close()
	}()
	buf, _ := io.ReadAll(r)
	return string(buf)
}

func TestLogOffByDefault(t *testing.T) {
	out := captureAllOutput(func() { testLog() })
	assert.Empty(t, out)
}

func TestLogLevels(t *testing.T) {
	stdlog.Default()
	var out strings.Builder
	warn := stdlog.New(&out, "<go-ceph>[WRN]", stdlog.Lshortfile)
	log.SetWarnf(warn.Printf)
	t.Run("Warnf", func(t *testing.T) {
		out.Reset()
		testLog()
		lines := strings.Split(out.String(), "\n")
		assert.Len(t, lines, 2)
		checkLines(t, lines)
	})
	debug := stdlog.New(&out, "<go-ceph>[DBG]", stdlog.Lshortfile)
	log.SetDebugf(debug.Printf)
	t.Run("Debugf", func(t *testing.T) {
		out.Reset()
		testLog()
		lines := strings.Split(out.String(), "\n")
		assert.Len(t, lines, 3)
		checkLines(t, lines)
	})
}
