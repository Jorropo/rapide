package rapide

import (
	"context"

	"github.com/Jorropo/rapide/internal"

	"github.com/ipfs/go-bitswap/network"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs-blockstore"
	"github.com/libp2p/go-libp2p-core/peer"

	"github.com/ipld/go-ipld-prime/codec"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/traversal/selector"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	// FIXME: remove import once sync.MapOf is added to std
	gmap "github.com/SaveTheRbtz/generic-sync-map-go"
)

var _ network.Receiver = (*Rapide)(nil)

type Rapide struct {
	bitswapPeerTasks gmap.MapOf[peer.ID, *bitswapPeerTaskSet]

	blockstore blockstore.Blockstore

	nodeLoader codec.Decoder
}

const returnChannelBufferSize = 256

// Get tries to download all blocks found under a selector.
// It runs async, the channel is closed when the request is finished.
// The context will be used to pass around tracing information to the blockstore and for cancellation.
func (r *Rapide) Get(ctx context.Context, root cid.Cid, future SelectorFuture, hints ...peer.ID) <-chan MaybeBlock {
	ctx, ctxCancel := context.WithCancel(ctx)
	ctx, span := internal.StartSpan(ctx, "GetBlock", trace.WithAttributes(attribute.String("Key", root.String())))
	c := make(chan MaybeBlock, returnChannelBufferSize)
	req := &request{
		rapide:    r,
		results:   c,
		ctx:       ctx,
		ctxCancel: ctxCancel,
		span:      span,
	}
	go req.contextWatchDog()
	go req.start(root, future, hints)
	return c
}

type MaybeBlock struct {
	// Err indicates and error, if it is set the block was not received.
	Err error

	// Block is only valid if Err == nil, it is the received block.
	Block blocks.Block
}

// SelectorFuture is a callback clients can give us to give us a selector later.
// This is needed because selectors need a seed node to start with.
type SelectorFuture func(n datamodel.Node) (selector.Selector, error)
