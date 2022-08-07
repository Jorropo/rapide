package rapide

import (
	"context"
	"fmt"
	"sync"

	prime "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/traversal/selector"

	"github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/libp2p/go-libp2p-core/peer"

	"go.opentelemetry.io/otel/trace"
)

// request is the event loop driven parallel traversal heart
type request struct {
	rapide *Rapide

	// anything touching the request must hold this lock, either as reading if it's partially compatible
	// or as writing if it's not
	lock sync.RWMutex

	results chan<- MaybeBlock

	// This is not used for cancellation of the request, this is used to pass the cancel to the blockstore.
	ctx       context.Context
	ctxCancel context.CancelFunc
	span      trace.Span
	done      bool // if done is true the request is over and all functions must exit asap

	selectorFuture SelectorFuture

	worklists []*worklist

	root *graphNode
}

func (r *request) start(root cid.Cid, future SelectorFuture, hints []peer.ID) {
	r.lock.Lock()
	defer r.lock.Unlock()

	b, err := r.rapide.blockstore.Get(r.ctx, root)
	if err == nil {
		// we have the block, awesome!

	} else if !ipld.IsNotFound(err) {
		// faulty datastore let's abort now
		r.closeWithErrorWhileHoldingGlobalLock(err)
		return
	} else {
		// we don't have the block let's fetch it
	}
	rg := &graphNode{cid: root}
	r.root = rg
}

func (r *request) contextWatchDog() {
	<-r.ctx.Done()
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.done {
		// we already closed, nothing more to do
		return
	}

	for _, w := range r.worklists {
		w.stopWithoutGraphCleanup()
	}
	r.closeWithErrorWhileHoldingGlobalLock(r.ctx.Err())
}

func (g *graphNode) expandBlock(b blocks.Block) error {
	g.lock.Lock()
	defer g.lock.Lock()

	if g.selector != nil {
		// TODO: add logging and ban badly behaving nodes
		// block already received
		return nil
	}

	if blockCid := b.Cid(); g.cid != blockCid {
		panic(fmt.Sprintf("inconsistent state, wrong block sent to the graph node, expected %q; got %q", g.cid, blockCid))
	}
	node, err := prime.Decode(b.RawData(), g.request.rapide.nodeLoader)

	if err != nil {
		return fmt.Errorf("failed to decode block into node: %w", err)
	}

	if g.parent == nil {
		// original root node
		g.selector, err = g.request.selectorFuture(node)
		if err != nil {
			return fmt.Errorf("failed to resolve root's future selector: %w", err)
		}
	} else {
		g.selector, err = g.parent.selector.Explore(node, g.oldSegment)
		if err != nil {
			return fmt.Errorf("failed to explore selector: %w", err)
		}
	}

	findLinks(g.selector, node, func())

}

type callbackWithBlock func(blocks.Block)

type graphNode struct {
	request *request
	parent  *graphNode

	lock sync.Mutex

	selector   selector.Selector
	oldSegment datamodel.PathSegment
	cid        cid.Cid

	subTasks []*graphNode
}

func (r *request) closeWithErrorWhileHoldingGlobalLock(err error) {
	r.results <- MaybeBlock{Err: err}
	close(r.results)
	r.done = true
	r.span.End()
}
