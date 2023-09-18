package loong

import (
	"context"
	"net"
	"net/http"
	"testing"

	"github.com/runner-mei/log/logtest"
)

func TestRunner(t *testing.T) {
	r := &Runner{
		Logger:             logtest.NewLogger(t),
		Network:            "http",
		ListenAt:           ":34456",
		CandidatePortStart: 50000,
		CandidatePortEnd:   60000,
	}
	net.Listen("tcp", r.ListenAt)

	ctx := context.Background()

	err := r.Start(ctx, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	}))
	if err != nil {
		t.Error(err)
		return
	}
	defer r.Stop(ctx)

	port, err := r.ListenPort()
	if err != nil {
		t.Error(err)
		return
	}

	if port == "34456" {
		t.Error("want 34456 got", port)
	}
}
