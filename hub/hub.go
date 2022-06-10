package hub

import (
	"context"
	"errors"
	"fmt"

	"github.com/hekmon/transmissionrpc/v2"
)

type Spoke interface {
	Query(context.Context, *transmissionrpc.Torrent) (string, error)
}

type Hub struct {
	spokes []Spoke
}

func NewHub(spokes []Spoke) *Hub {
	return &Hub{
		spokes: spokes,
	}
}

func (h *Hub) Query(ctx context.Context, t *transmissionrpc.Torrent) (string, error) {
	for i, spoke := range h.spokes {
		resp, err := spoke.Query(ctx, t)
		if err != nil {
			return "", fmt.Errorf("hub error: spoke[%d] failed: %w", i, err)
		}
		if resp != "" {
			return resp, nil
		}
	}
	return "", errors.New("hub error: no spoke provided result")
}
