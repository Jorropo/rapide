package rapide

import (
	"context"
	"sync"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/traversal/selector"
	"github.com/libp2p/go-libp2p-core/peer"
	"go.opentelemetry.io/otel/trace"
)

// request is the event loop driven parallel traversal heart
type request struct {
	rapide *Rapide

	lock sync.Mutex

	results chan<- MaybeBlock

	// This is not used for cancellation of the request, this is used to pass the cancel to the blockstore.
	ctx       context.Context
	ctxCancel context.CancelFunc
	span      trace.Span

	selectorFuture SelectorFuture

	worklists []*worklist
}

func (r *request) start(root cid.Cid, future SelectorFuture, hints []peer.ID) {
}

func (r *request) cancel() {
	r.ctxCancel()
	r.lock.Lock()
	defer r.lock.Unlock()
	for _, w := range r.worklists {
		w.stop()
	}
	r.span.End()
}

type callbackWithBlock func(blocks.Block)

type graphNode struct {
	request *request

	lock sync.Mutex

	selector selector.Selector
}
