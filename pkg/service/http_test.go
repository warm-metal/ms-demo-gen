package service

import (
	"context"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())
	os.Exit(m.Run())
}

func httpTraffic(t *testing.T, payloadSize, uploadSize int, queryInParallel bool) {
	randomSize := func(max int) int {
		if max == 0 {
			return max
		}

		return rand.Intn(max)
	}

	optServer := &Options{
		Address:         ":8001",
		PayloadSize:     randomSize(payloadSize),
		UploadSize:      randomSize(uploadSize),
		QueryInParallel: queryInParallel,
	}

	optServer2 := &Options{
		Address:         ":8002",
		PayloadSize:     randomSize(payloadSize),
		UploadSize:      randomSize(uploadSize),
		QueryInParallel: queryInParallel,
	}

	optClient := &Options{
		Address:         ":8000",
		Upstream:        []string{optServer.Address, optServer2.Address},
		PayloadSize:     randomSize(payloadSize),
		UploadSize:      randomSize(uploadSize),
		QueryInParallel: queryInParallel,
	}

	client := CreateServer(optClient)
	server := CreateServer(optServer)
	server2 := CreateServer(optServer2)

	ctx, cancel := context.WithCancel(context.Background())
	clientDone := client.LoopInBackground(ctx)
	serverDone := server.LoopInBackground(ctx)
	server2Done := server2.LoopInBackground(ctx)
	defer func() {
		cancel()
		<-serverDone
		<-server2Done
		<-clientDone
	}()

	resp, err := http.Get("http://" + optClient.Address)
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
	if len(body.String()) != optClient.PayloadSize {
		t.Logf("client payload size: %d, server1 payload size: %d, server2 payload size:%d\n",
			optClient.PayloadSize, optServer.PayloadSize, optServer2.PayloadSize)
		t.FailNow()
	}

	if optClient.PayloadSize > optServer.PayloadSize+optServer2.PayloadSize &&
		(!strings.Contains(body.String(), optServer.Address) || !strings.Contains(body.String(), optServer2.Address)) {
		t.Logf("client payload size: %d, server1 payload size: %d, server2 payload size:%d\n",
			optClient.PayloadSize, optServer.PayloadSize, optServer2.PayloadSize)
		t.FailNow()
	}
}

func TestHttpTrafficWoPayloads(t *testing.T) {
	httpTraffic(t, 0, 0, false)
	httpTraffic(t, 0, 0, true)
}

func TestHttpTrafficWPayloads(t *testing.T) {
	httpTraffic(t, 512, 64, false)
	httpTraffic(t, 512, 64, true)
}
