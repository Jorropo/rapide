package rapide

import (
	"runtime"
	"unsafe"
)

func comparePointers[T *any](a, b T) bool {
	r := uintptr(unsafe.Pointer(a)) == uintptr(unsafe.Pointer(b))
	// KeepAlive to avoid ABA bugs because we are using pointers as identity.
	// This prevent a race with the GC & allocator by keeping the objects alive.
	runtime.KeepAlive(a)
	runtime.KeepAlive(b)
	return r
}
