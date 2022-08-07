package rapide

import (
	"sync"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/peer"
)

type bitswapPeerTaskSet struct {
	lock sync.Mutex

	tasks map[cid.Cid]ask

	// if removed is true, the wantlist was empty and you need to add it back
	removed bool
}

func (r *Rapide) askForMoreBlocks(p peer.ID, wanted []cidPriorityPair, callback callbackWithBlock) error {
	pt, _ := r.bitswapPeerTasks.LoadOrStore(p, &bitswapPeerTaskSet{})

tryPt:
	pt.lock.Lock()

	// did we finished while we were taking the lock ?
	if pt.removed {
		newPt, loaded := r.bitswapPeerTasks.LoadOrStore(p, pt)
		if loaded {
			// someone else already made a new pt
			pt.lock.Unlock()
			pt = newPt
			goto tryPt
		}
	}

	blocksToGet := make([]cid.Cid, len(wanted))
	var newBlocks int
	for _, want := range wanted {
		ask, alreadyAsked := pt.tasks[want.cid]
		if alreadyAsked {
			if ask.priority > want.priority {

			}
		}
	}
}

type cidPriorityPair struct {
	cid      cid.Cid
	priority uint32
}

type ask struct {
	priority uint32

	callbacks []callbackWithBlock
}
