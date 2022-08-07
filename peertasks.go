package rapide

import (
	"sync"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/peer"
)

type bitswapPeerTaskSet struct {
	rapide *Rapide

	lock sync.Mutex

	tasks []multiAsk

	worklists uint64

	self peer.ID
}

func (r *Rapide) getNewWorklistForPeer(p peer.ID) *worklist {
	pt, _ := r.bitswapPeerTasks.LoadOrStore(p, &bitswapPeerTaskSet{rapide: r, self: p})

tryPt:
	pt.lock.Lock()

	// did we finished while we were taking the lock ?
	if pt.worklists == 0 {
		newPt, loaded := r.bitswapPeerTasks.LoadOrStore(p, pt)
		if loaded {
			// someone else already made a new pt
			pt.lock.Unlock()
			pt = newPt
			goto tryPt
		}
	}
	pt.worklists++

	return &worklist{
		peer: pt,
	}
}

type multiAsk struct {
	cid cid.Cid

	// worklists is sorted by priority
	worklists []multiAskWorklistItem
}

type multiAskWorklistItem struct {
	worklist *worklist
	priority uint32
}

type ask struct {
	cid      cid.Cid
	priority uint32
}

type worklist struct {
	peer *bitswapPeerTaskSet

	asks []ask
}

func (w *worklist) stop() {
	pt := w.peer
	pt.lock.Lock()
	defer pt.lock.Unlock()

	if pt.worklists == 1 {
		// fast path for when we are the last worklist
		pt.rapide.bitswapPeerTasks.Delete(pt.self)
		pt.worklists = 0
		pt.tasks = nil // don't keep the capacity to avoid holding a reference to the items and worklist
		return
	}

	var j int
	for _, a := range w.asks {
		for a.cid != pt.tasks[j].cid {
			j++
		}

		for i, item := range pt.tasks[j].worklists {
			if comparePointers(w, item.worklist) {
				pt.tasks[j].worklists = removeItem(pt.tasks[j].worklists, i)
				goto afterWorklistRemove
			}
		}
		panic("worklist not found in pt.tasks[j]")
	afterWorklistRemove:
	}
	pt.worklists--
}
