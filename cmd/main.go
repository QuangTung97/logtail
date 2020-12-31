package main

import (
	"fmt"
	"logtail"
	"os"
	"path"
)

func main() {
	if len(os.Args) <= 1 {
		panic("missing 'filepath'")
	}

	filepath := os.Args[1]

	dirname := path.Dir(filepath)
	filename := path.Base(filepath)

	watcher := logtail.NewWatcher(dirname)

	go func() {
		watcher.Watch()
	}()

	logtail.TailFile(dirname, filename, watcher.GetWatchChan(), func(line string) {
		fmt.Print("LINE: ", line)
	})
}
