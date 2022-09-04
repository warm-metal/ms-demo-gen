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

func httpTraffic(t testing.TB, payloadSize, uploadSize int, queryInParallel bool, loop int) {
	randomSize := func(max int) int {
		if max == 0 {
			return max
		}

		return rand.Intn(max)
	}

	optServer := &Options{
		Address:         "127.0.0.1:8001",
		PayloadSize:     randomSize(payloadSize),
		UploadSize:      randomSize(uploadSize),
		QueryInParallel: queryInParallel,
	}

	optServer2 := &Options{
		Address:         "127.0.0.1:8002",
		PayloadSize:     randomSize(payloadSize),
		UploadSize:      randomSize(uploadSize),
		QueryInParallel: queryInParallel,
	}

	optClient := &Options{
		Address:         "127.0.0.1:8000",
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

	for i := 0; i < loop; i++ {
		resp, err := http.Get("http://" + optClient.Address)
		if err != nil {
			t.Log(err)
			t.FailNow()
			return
		}

		if resp.StatusCode != http.StatusOK {
			t.Log(resp.StatusCode)
			t.FailNow()
			return
		}

		body := &strings.Builder{}
		io.Copy(body, resp.Body)
		resp.Body.Close()
		if len(body.String()) != optClient.PayloadSize {
			t.Logf("client payload size: %d, server1 payload size: %d, server2 payload size:%d, response size:%d\n",
				optClient.PayloadSize, optServer.PayloadSize, optServer2.PayloadSize, len(body.String()))
			t.FailNow()
		}
	}
}

func TestHttpTrafficWoPayloads(t *testing.T) {
	httpTraffic(t, 0, 0, false, 1)
	httpTraffic(t, 0, 0, true, 1)
}

func TestHttpTrafficWPayloads(t *testing.T) {
	httpTraffic(t, 512, 64, false, 1)
	httpTraffic(t, 512, 64, true, 1)
}

func BenchmarkHttpService(b *testing.B) {
	httpTraffic(b, 512, 64, true, 1000)
}

func BenchmarkNonDataHttpService(b *testing.B) {
	httpTraffic(b, 0, 0, true, 1000)
}
