package service

import (
	"net/http"
	"strings"
)

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

	

	// delta in plans > 0
	firstOversizedPos := 0
	// delta in plans < 0
	firstUndersizedPos := 0
	for i, p := range payloadSizePlan {
		if p == 0 {
			continue
		}



		if p < 0 {
			// find an oversized position from the begining then balance it
			for j := firstOversizedPos; j < i; j++ {
				if payloadSizePlan[j] <= 0 {
					firstOversizedPos++
					continue
				}

				payloadSizePlan[j] += payloadSizePlan[i]
				payloadSizePlan[i] = 0
				if payloadSizePlan[j] >= 0 {
					break
				}

				payloadSizePlan[i] = -payloadSizePlan[j]
				payloadSizePlan[j] = 0
				firstOversizedPos++
			}
			continue
		}

		// if p > 0, then find an undersized position from the beginning
		for j := firstUndersizedPos; j < i; j++ {
			if payloadSizePlan[j] > 0 {
				firstUndersizedPos++
				continue
			}

			payloadSizePlan[j] += payloadSizePlan[i]
			payloadSizePlan[i] = 0
			if payloadSizePlan[j] <= 0 {
				break
			}

			payloadSizePlan[i] = -payloadSizePlan[j]
			payloadSizePlan[j] = 0
			firstUndersizedPos++
		}
	}

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
	return nil
}

func (c *RemoteClient) syncQuery(upstream string) string {
	return ""
}

func (c *RemoteClient) asyncQuery(upstream string, out chan string) {
	defer close(out)
	resp := c.syncQuery(upstream)
	out <- resp
}
