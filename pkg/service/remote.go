package service

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	rands "github.com/xyproto/randomstring"
)

type SignBit int

func (b SignBit) Counter() SignBit {
	return 1 - b
}

const (
	SignBitPositive = SignBit(iota)
	SignBitNegative
)

func GetSignBit(i int) SignBit {
	if math.Signbit(float64(i)) {
		return SignBitNegative
	}

	return SignBitPositive
}

func buildPayloadSizePlan(payloadSizePlan []int) {
	focusedPos := [2]int{0, 0}
	for i, p := range payloadSizePlan {
		if p == 0 {
			continue
		}

		sign := GetSignBit(p)
		counterSign := sign.Counter()
		for j := focusedPos[counterSign]; j < i; j++ {
			focusedSign := GetSignBit(payloadSizePlan[j])
			if payloadSizePlan[j] == 0 || sign == focusedSign {
				focusedPos[counterSign] += 1
				continue
			}

			payloadSizePlan[j] += payloadSizePlan[i]
			payloadSizePlan[i] = 0
			if focusedSign == GetSignBit(payloadSizePlan[j]) {
				break
			}

			payloadSizePlan[i] = payloadSizePlan[j]
			payloadSizePlan[j] = 0
			focusedPos[counterSign] += 1
		}
	}
}

func fillStringBuilderOrDie(b *strings.Builder, vs ...string) {
	for _, v := range vs {
		if _, err := b.WriteString(v); err != nil {
			panic(err)
		}
	}
}

type RemoteClient struct {
	upstream       []string
	inParallel     bool
	longConnection bool
	timeout        time.Duration
}

func (c *RemoteClient) call(uploadSize int, respWriter io.Writer) (waitingList []chan io.Writer) {
	if len(c.upstream) == 0 {
		return
	}

	waitingList = make([]chan io.Writer, len(c.upstream))
	for i, up := range c.upstream {
		waitingList[i] = make(chan io.Writer, 1)
		w := respWriter
		if w == nil {
			w = &strings.Builder{}
		}
		if c.inParallel {
			go c.asyncQuery(uploadSize, up, w, waitingList[i])
		} else {
			c.syncQuery(uploadSize, up, w)
			waitingList[i] <- w
		}
	}

	return
}

func (c *RemoteClient) Discard(uploadSize int) {
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		panic(err)
	}

	defer devNull.Close()
	waitingList := c.call(uploadSize, devNull)
	for _, w := range waitingList {
		<-w
	}
}

func (c *RemoteClient) Query(uploadSize, maxPayloadSize int) (resp string) {
	waitingList := c.call(uploadSize, nil)
	payloadSizePlan := make([]int, len(c.upstream))
	resps := make([]string, len(c.upstream))
	maxPayloadSizePerUpstream := maxPayloadSize / len(c.upstream)
	for i, w := range waitingList {
		reader := <-w
		resps[i] = reader.(*strings.Builder).String()
		// taking length of the upsteam url into account
		if maxPayloadSize > 0 {
			payloadSizePlan[i] = len(resps[i]) + len(c.upstream[i]) + 1 - maxPayloadSizePerUpstream
		}
	}

	if maxPayloadSize == 0 {
		return ""
	}

	buildPayloadSizePlan(payloadSizePlan)

	b := strings.Builder{}
	if maxPayloadSize > 0 {
		b.Grow(maxPayloadSize)
	}

	for i, resp := range resps {
		if payloadSizePlan[i] > 0 {
			if len(resp) >= payloadSizePlan[i] {
				fillStringBuilderOrDie(&b, c.upstream[i], ":")
				fillStringBuilderOrDie(&b, resp[:len(resp)-payloadSizePlan[i]])
			}
		} else {
			fillStringBuilderOrDie(&b, c.upstream[i], ":")
			fillStringBuilderOrDie(&b, resp)
		}
	}
	return b.String()
}

func (c *RemoteClient) genClient(upstream string) *http.Client {
	return &http.Client{
		Timeout: c.timeout,
		Transport: &http.Transport{
			DisableKeepAlives: !c.longConnection,
		},
	}
}

func (c *RemoteClient) syncQuery(uploadSize int, upstream string, respWriter io.Writer) {
	client := c.genClient(upstream)
	url := "http://" + upstream
	var err error
	var resp *http.Response
	if uploadSize > 0 {
		resp, err = client.Post(url, "text/plain", strings.NewReader(rands.HumanFriendlyEnglishString(uploadSize)))
	} else {
		resp, err = client.Get(url)
	}

	if err != nil {
		respWriter.Write([]byte(fmt.Sprintf("Error: %s", err)))
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respWriter.Write([]byte(fmt.Sprintf("Error: %s", resp.Status)))
		return
	}

	io.Copy(respWriter, resp.Body)
}

func (c *RemoteClient) asyncQuery(uploadSize int, upstream string, respWriter io.Writer, out chan io.Writer) {
	defer close(out)
	c.syncQuery(uploadSize, upstream, respWriter)
	out <- respWriter
}

func NewClient(opts *Options) *RemoteClient {
	return &RemoteClient{
		upstream:       opts.Upstream,
		inParallel:     opts.QueryInParallel,
		longConnection: opts.LongConn,
		timeout:        opts.Timeout,
	}
}
