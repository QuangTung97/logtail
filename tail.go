package logtail

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
)

func batchCheckWatchChan(filename string, ch <-chan WatchEvent) (isCreate bool) {
	isCreate = false
	count := 1

	first := <-ch
	if first.Filename == filename && first.Type == EventTypeCreate {
		isCreate = true
	}

	for {
		select {
		case event := <-ch:
			count++
			if event.Filename == filename && event.Type == EventTypeCreate {
				isCreate = true
			}

		default:
			fmt.Println(count)
			return
		}
	}
}

// TailFile ...
func TailFile(dirname string, filename string, ch <-chan WatchEvent, fn func(line string)) {
	filepath := path.Join(dirname, filename)

	file, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(file)

ReadLoop:
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			isCreate := batchCheckWatchChan(filename, ch)
			if isCreate {
				_ = file.Close()

				newFile, err := os.Open(filepath)
				if err != nil {
					panic(err)
				}

				file = newFile
				reader = bufio.NewReader(file)

				continue ReadLoop
			}

			currentPos, err := file.Seek(0, os.SEEK_CUR)
			if err != nil {
				panic(err)
			}

			stat, err := file.Stat()
			if err != nil {
				panic(err)
			}

			if currentPos > stat.Size() {
				_, err := file.Seek(0, os.SEEK_SET)
				if err != nil {
					panic(err)
				}
			}

			continue ReadLoop
		}
		if err != nil {
			panic(err)
		}

		fn(line)
	}
}
