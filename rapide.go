package rapide

import (
	"github.com/libp2p/go-libp2p-core/peer"

	"github.com/ipfs/go-bitswap/network"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/traversal/selector"

	// FIXME: remove import once sync.MapOf is added to std
	gmap "github.com/SaveTheRbtz/generic-sync-map-go"
)

var _ network.Receiver = (*Rapide)(nil)

type Rapide struct {
	bitswapPeerTasks gmap.MapOf[peer.ID, *bitswapPeerTaskSet]
}

const returnChannelBufferSize = 256

// Get tries to download all blocks found under a selector.
// It runs async, the channel is closed when the request is finished.
// It also return a cancel function instead of using context, that because I choose
// to write an event loop instead of using goroutines. I know go's scheduler is so good
// I am supposed to just use goroutines, but my brain made more sense of the event loop
// version so I use an event loop OK ? This most probably saves lots of ram cuz I use 1 or 2 pointers
// where goroutines would need multiple channels and stacks. No I havn't benchmarked that claim.
// The cancel function is synchronous and waits until it is fully closed to return.
func (r *Rapide) Get(root cid.Cid, future SelectorFuture, hints ...peer.ID) (<-chan MaybeBlock, func()) {
	c := make(chan MaybeBlock, returnChannelBufferSize)
	req := request{
		rapide:  r,
		results: c,
	}
	go req.start(root, future, hints)
	return c, req.cancel
}

type MaybeBlock struct {
	// Key is always set to the corresponding value found in the traversal
	Key cid.Cid

	// Err indicates and error, if it is set the block was not received.
	Err error

	// Block is only valid if Err == nil, it is the received block.
	Block blocks.Block
}

// SelectorFuture is a callback clients can give us to give us a selector later.
// This is needed because selectors need a seed node to start with.
type SelectorFuture func(n datamodel.Node) (selector.Selector, error)
