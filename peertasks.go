package rapide

import (
	"sync"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/peer"
)

type bitswapPeerTaskSet struct {
	rapide *Rapide

	lock sync.Mutex

	worklists []*worklist
	tasks     []multiAsk

	touched         uint64
	lastTouchedSent uint64

	self peer.ID
}

func (r *Rapide) getNewWorklistForPeer(p peer.ID) *worklist {
	pt, _ := r.bitswapPeerTasks.LoadOrStore(p, &bitswapPeerTaskSet{rapide: r, self: p})

tryPt:
	pt.lock.Lock()

	// did we finished while we were taking the lock ?
	if len(pt.worklists) == 0 {
		newPt, loaded := r.bitswapPeerTasks.LoadOrStore(p, pt)
		if loaded {
			// someone else already made a new pt
			pt.lock.Unlock()
			pt = newPt
			goto tryPt
		}
	}

	w := &worklist{
		peer: pt,
	}

	pt.worklists = append(pt.worklists, w)

	return w
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

type worklist struct {
	peer *bitswapPeerTaskSet

	asks []cid.Cid
}

func (w *worklist) stopWithoutGraphCleanup() {
	pt := w.peer
	pt.lock.Lock()
	defer pt.lock.Unlock()

	var needToSendFull bool
	if len(pt.worklists) == 1 {
		// fast path for when we are the last worklist
		pt.rapide.bitswapPeerTasks.Delete(pt.self)
		pt.worklists = nil
		needToSendFull = len(pt.tasks) != 0
		pt.tasks = nil // don't keep the capacity to avoid holding a reference to the items and worklist
	} else {
		var c int
		var j int
		tasks := pt.tasks
	AsksLoop:
		for _, a := range w.asks {
			for a != tasks[j].cid {
				tasks[c] = tasks[j]
				j++
				c++
			}

			for i, item := range tasks[j].worklists {
				if comparePointers(w, item.worklist) {
					if len(tasks[j].worklists) > 1 {
						// an other worklist is waiting on that item let's keep it and remove us from the worklists
						tasks[j].worklists = removeItem(tasks[j].worklists, i)
						c++
					}
					j++
					continue AsksLoop
				}
			}
			panic("worklist not found in tasks[j]")
		}
		if j != c {
			// copy remaining unscanned items
			c += copy(tasks[c:], tasks[j:])
			// memclr unused part of tasks to avoid holding references
			for i := range pt.tasks[c:] {
				tasks[i] = multiAsk{}
			}
			pt.tasks = tasks[:c]
			needToSendFull = true
		}
		for i, z := range pt.worklists {
			if comparePointers(w, z) {
				pt.worklists = removeItem(pt.worklists, i)
				goto WorklistRemoved
			}
		}
		panic("worklist not found in pt.worklists")
	WorklistRemoved:
	}

	if needToSendFull {
		pt.touched++
		go pt.sendFull()
	}
}

func (pt *bitswapPeerTaskSet) sendFull() {
	pt.lock.Lock()
	defer pt.lock.Unlock()

	if pt.touched == pt.lastTouchedSent {
		return
	}

	panic("not implemented")
}
