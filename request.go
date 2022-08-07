package rapide

import (
	"context"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/peer"

	"go.opentelemetry.io/otel/trace"
)

// request is the event loop driven parallel traversal heart
type request struct {
	rapide *Rapide

	ctx     context.Context
	span    trace.Span
	results chan<- MaybeBlock
}

func (r *request) start(root cid.Cid, future SelectorFuture, hints []peer.ID) {

}

// done is like wg.Done but following the event loop
func (r *request) done() {

}

type callbackWithBlock func(blocks.Block)
