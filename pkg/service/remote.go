package service

import (
	"fmt"
	"io"
	"math"
	"net/http"
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
				focusedPos[counterSign]+=1
				continue
			}

			payloadSizePlan[j] += payloadSizePlan[i]
			payloadSizePlan[i] = 0
			if focusedSign == GetSignBit(payloadSizePlan[j]) {
				break
			}

			payloadSizePlan[i] = payloadSizePlan[j]
			payloadSizePlan[j] = 0
			focusedPos[counterSign]+=1
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
}

func (c *RemoteClient) Query(maxPayloadSize int) (resp string) {
	if len(c.upstream) == 0 {
		return
	}

	waitingList := make([]chan string, len(c.upstream))
	for i, up := range c.upstream {
		waitingList[i] = make(chan string, 1)
		if c.inParallel {
			go c.asyncQuery(up, waitingList[i])
		} else {
			waitingList[i] <- c.syncQuery(up)
		}
	}

	maxPayloadSizePerUpstream := maxPayloadSize / len(c.upstream)
	payloadSizePlan := make([]int, len(c.upstream))
	resps := make([]string, len(c.upstream))
	for i, w := range waitingList {
		resps[i] = <-w
		// taking length of the upsteam url into account
		payloadSizePlan[i] = len(resps[i]) + len(c.upstream[i]) + 1 - maxPayloadSizePerUpstream
	}

	buildPayloadSizePlan(payloadSizePlan)

	b := strings.Builder{}
	b.Grow(maxPayloadSize)
	for i, resp := range resps {
		fillStringBuilderOrDie(&b, c.upstream[i], ":")
		if payloadSizePlan[i] > 0 {
			fillStringBuilderOrDie(&b, resp[:len(resp)-payloadSizePlan[i]])
		} else {
			fillStringBuilderOrDie(&b, resp)
		}
	}
	return b.String()
}

func (c *RemoteClient) genClient(upstream string) *http.Client {
	return http.DefaultClient
}

func (c *RemoteClient) syncQuery(upstream string) string {
	client := c.genClient(upstream)
	resp, err := client.Get("http://"+upstream)
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("Error: %s", resp.Status)
	}

	if c.longConnection {
		defer resp.Body.Close()
	}

	b := &strings.Builder{}
	io.Copy(b, resp.Body)
	return b.String()
}

func (c *RemoteClient) asyncQuery(upstream string, out chan string) {
	defer close(out)
	resp := c.syncQuery(upstream)
	out <- resp
}
