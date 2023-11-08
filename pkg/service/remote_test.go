package service

import (
	"math/rand"
	"net"
	"net/http"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestPayloadSizePlanningWithAllBalanced(t *testing.T) {
	balanced := []int{0, 0, 0, 0, 0, 0}
	buildPayloadSizePlan(balanced)
	for _, v := range balanced {
		if v != 0 {
			t.FailNow()
		}
	}
}

func TestPayloadSizePlanningWithAllOversized(t *testing.T) {
	plan := []int{1, 2, 3, 4, 5, 6}
	buildPayloadSizePlan(plan)
	for i := range plan {
		if plan[i] != i+1 {
			t.FailNow()
		}
	}

	sort.Sort(sort.Reverse(sort.IntSlice(plan)))
	for i := range plan {
		if plan[i] != 6-i {
			t.FailNow()
		}
	}
}

func TestPayloadSizePlanningWithAllUndersized(t *testing.T) {
	plan := []int{-1, -2, -3, -4, -5, -6}
	buildPayloadSizePlan(plan)
	for i := range plan {
		if plan[i] != -i-1 {
			t.FailNow()
		}
	}
}

func TestBalancedPayloadSizePlanning(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	plan := []int{6, 3, -1, -4, 0, 2, -6}

	for i := 0; i < 5; i++ {
		rand.Shuffle(len(plan), func(i, j int) { plan[i], plan[j] = plan[j], plan[i] })
		buildPayloadSizePlan(plan)
		for _, v := range plan {
			if v != 0 {
				t.Log(plan)
				t.FailNow()
			}
		}
	}
}

func TestOversizedPayloadSizePlanning(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	plan := []int{6, 3, -3, -1, 0, 2, -1}
	sum := func(a []int) int {
		s := 0
		for i := range a {
			s += a[i]
		}

		return s
	}

	for i := 0; i < 5; i++ {
		rand.Shuffle(len(plan), func(i, j int) { plan[i], plan[j] = plan[j], plan[i] })
		s := sum(plan)
		buildPayloadSizePlan(plan)
		if s != sum(plan) {
			t.Log(plan)
			t.FailNow()
		}
	}
}

func BenchmarkRemoteQuery(b *testing.B) {
	serverMux := &http.ServeMux{}
	payload := []byte(strings.Repeat("X", 512))
	serverMux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		resp.Write(payload)
	})

	server := http.Server{
		Addr:    "127.0.0.1:8000",
		Handler: serverMux,
	}

	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		panic(err)
	}

	go server.Serve(ln)
	defer server.Close()

	cli := NewClient(&Options{
		Upstream: []string{server.Addr},
	})

	for i := 0; i < 100000; i++ {
		cli.Query(nil, nil, 512)
	}
}
