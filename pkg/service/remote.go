package service

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
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
	upstream   []string
	inParallel bool
	client     http.Client
}

func (c *RemoteClient) call(headers map[string]string, uploadReader *strings.Reader, respWriter io.Writer) (waitingList []chan io.Writer) {
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
			go c.asyncQuery(headers, uploadReader, up, w, waitingList[i])
		} else {
			c.syncQuery(headers, uploadReader, up, w)
			waitingList[i] <- w
		}
	}

	return
}

func (c *RemoteClient) Discard(headers map[string]string, uploadReader *strings.Reader) {
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		panic(err)
	}

	defer devNull.Close()
	waitingList := c.call(headers, uploadReader, devNull)
	for _, w := range waitingList {
		<-w
	}
}

func (c *RemoteClient) Query(headers map[string]string, uploadReader *strings.Reader, maxPayloadSize int) (resp string) {
	waitingList := c.call(headers, uploadReader, nil)
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

func (c *RemoteClient) syncQuery(headers map[string]string, uploadReader *strings.Reader, upstream string, respWriter io.Writer) {
	url := "http://" + upstream
	var err error
	var resp *http.Response
	var req *http.Request
	if uploadReader != nil {
		req, err = http.NewRequest("POST", url, uploadReader)
	} else {
		req, err = http.NewRequest("GET", url, nil)
	}

	if err != nil {
		respWriter.Write([]byte(fmt.Sprintf("Error: %s", err)))
		return
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp, err = c.client.Do(req)
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

func (c *RemoteClient) asyncQuery(headers map[string]string, uploadReader *strings.Reader, upstream string, respWriter io.Writer, out chan io.Writer) {
	defer close(out)
	c.syncQuery(headers, uploadReader, upstream, respWriter)
	out <- respWriter
}

func NewClient(opts *Options) RemoteClient {
	return RemoteClient{
		upstream:   opts.Upstream,
		inParallel: opts.QueryInParallel,
		client: http.Client{
			Timeout: opts.Timeout,
			Transport: &http.Transport{
				DisableKeepAlives: !opts.LongConn,
			},
		},
	}
}
