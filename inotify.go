package logtail

import (
	"reflect"
	"syscall"
	"unsafe"
)

func cStringLen(s []byte) int {
	for i, e := range s {
		if e == 0 {
			return i
		}
	}
	return 0
}

// EventType ...
type EventType int

const (
	// EventTypeCreate ...
	EventTypeCreate EventType = 1
	// EventTypeModify ...
	EventTypeModify EventType = 2
)

// WatchEvent ...
type WatchEvent struct {
	Filename string
	Type     EventType
}

// DirectoryWatcher ...
type DirectoryWatcher struct {
	inotifyFd int
	watchDesc int
	watchChan chan WatchEvent
}

// NewWatcher ...
func NewWatcher(dirname string) *DirectoryWatcher {
	inotifyFd, err := syscall.InotifyInit()
	if err != nil {
		panic(err)
	}

	watchDesc, err := syscall.InotifyAddWatch(
		inotifyFd, dirname, syscall.IN_MODIFY|syscall.IN_CREATE)
	if err != nil {
		panic(err)
	}

	return &DirectoryWatcher{
		inotifyFd: inotifyFd,
		watchDesc: watchDesc,
		watchChan: make(chan WatchEvent),
	}
}

func eventTypeFromInotify(mask uint32) EventType {
	if mask&syscall.IN_CREATE != 0 {
		return EventTypeCreate
	}
	return EventTypeModify
}

// Watch ...
func (w *DirectoryWatcher) Watch() {
	buffer := make([]byte, syscall.SizeofInotifyEvent*128)
	for {
		numReads, err := syscall.Read(w.inotifyFd, buffer)
		if err != nil {
			panic(err)
		}

		for offset := 0; offset+syscall.SizeofInotifyEvent < numReads; {
			event := (*syscall.InotifyEvent)(unsafe.Pointer(&buffer[offset]))

			if event.Wd == int32(w.watchDesc) {
				p := unsafe.Pointer(&event.Name)

				header := reflect.SliceHeader{
					Data: uintptr(p),
					Len:  int(event.Len),
					Cap:  int(event.Len),
				}

				array := *(*[]byte)(unsafe.Pointer(&header))
				name := string(array[:cStringLen(array)])

				w.watchChan <- WatchEvent{
					Filename: name,
					Type:     eventTypeFromInotify(event.Mask),
				}
			}

			offset += int(syscall.SizeofInotifyEvent + event.Len)
		}
	}
}

// GetWatchChan ...
func (w *DirectoryWatcher) GetWatchChan() <-chan WatchEvent {
	return w.watchChan
}

// Shutdown ...
func (w *DirectoryWatcher) Shutdown() {
	err := syscall.Close(w.inotifyFd)
	if err != nil {
		panic(err)
	}
	close(w.watchChan)
}
