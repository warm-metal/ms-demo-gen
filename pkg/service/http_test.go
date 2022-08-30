package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestHttpTraffic(t *testing.T) {
	optServer := &Options{
		Address: ":8001",
	}

	optClient := &Options{
		Address: ":8000",
		Upstream: []string{optServer.Address},
	}

	client := CreateServer(optClient)
	server := CreateServer(optServer)

	ctx, cancel := context.WithCancel(context.Background())
	clientDone := client.LoopInBackground(ctx)
	serverDone := server.LoopInBackground(ctx)
	defer func() {
		cancel()
		<-serverDone
		<-clientDone
	}()

	resp, err := http.Get("http://"+optClient.Address)
	if err != nil {
		t.Log(err)
		t.FailNow()
		return
	}

	if resp.StatusCode != http.StatusOK {
		t.Log(err)
		t.FailNow()
		return
	}

	defer resp.Body.Close()
	body := &strings.Builder{}
	io.Copy(body, resp.Body)
	t.Log(body.String())
}
