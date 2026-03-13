//Imporant NOTE:still an issue with a short write on using this patter should be resolved soon
// Storage Engine
package storage

import (
	"bufio"
	"log"
	lg "log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"mouloud.com/thor/utils"
)

type WorkerInput struct {
	errChan chan<- error
	log     string
	resChan chan<- string
}

func NewWorkerInput(errChan chan<- error, log string, resChan chan<- string) WorkerInput {
	return WorkerInput{errChan: errChan, log: log, resChan: resChan}
}

type Engine struct {
	Dir           string
	SegmentSize   int64
	SegmentsCount int
	numWorkers    int
	WorkersChan   chan WorkerInput
	mu            sync.RWMutex
	currentFile   *os.File
	currentWriter *bufio.Writer
}

func NewEngine(dir string, segmentSize int64, numWorkers int) (*Engine, error) {
	workChan := make(chan WorkerInput, numWorkers)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	segmentCount := 0
	for {
		fileName := dir + "/segment_" + strconv.Itoa(segmentCount) + ".log"
		stat, err := os.Stat(fileName)
		if os.IsNotExist(err) {
			break
		}
		if stat.Size() < segmentSize {
			break
		}
		segmentCount++
	}

	e := &Engine{
		Dir:           dir,
		SegmentSize:   segmentSize,
		SegmentsCount: segmentCount,
		numWorkers:    numWorkers,
		WorkersChan:   workChan,
	}

	// Open the current segment file on startup
	if err := e.openCurrentSegment(); err != nil {
		return nil, err
	}

	return e, nil
}

// openCurrentSegment opens the current segment file and sets up the persistent writer.
// Must be called with write lock held, except during initialization.
func (e *Engine) openCurrentSegment() error {
	fileName := e.GetFileName(e.SegmentsCount)
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	e.currentFile = file
	e.currentWriter = bufio.NewWriterSize(file, 256*1024) // 256kb buffer
	return nil
}

// rotateSegment flushes and closes the current segment, then opens a new one.
// Must be called with write lock held.
func (e *Engine) rotateSegment() error {
	if e.currentWriter != nil {
		if err := e.currentWriter.Flush(); err != nil {
			return err
		}
	}
	if e.currentFile != nil {
		if err := e.currentFile.Sync(); err != nil {
			return err
		}
		if err := e.currentFile.Close(); err != nil {
			return err
		}
	}
	e.SegmentsCount++
	lg.Printf("Rotating to new segment: %s\n", e.GetFileName(e.SegmentsCount))
	return e.openCurrentSegment()
}

func (e *Engine) startSyncTicker() {
	ticker := time.NewTicker(200 * time.Millisecond)
	go func() {
		for range ticker.C {
			e.mu.RLock()
			file := e.currentFile
			writer := e.currentWriter
			e.mu.RUnlock()
			if writer != nil {
				e.mu.Lock()
				_ = e.currentWriter.Flush()
				e.mu.Unlock()
			}
			if file != nil {
				_ = file.Sync()
			}
		}
	}()
}

func (e *Engine) StartWorkers() {
	// Watcher goroutine
	go func() {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			lg.Fatalln("Failed to start watcher", err.Error())
			return
		}

		go func() {

			debounced:=utils.New(250*time.Millisecond)
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					if event.Has(fsnotify.Write) {
						e.mu.RLock()
						count := e.SegmentsCount
						e.mu.RUnlock()
						
						debounced(func() {
							lg.Println("Checkig segment")
if e.shouldCreateNewSegment(count) {
							e.mu.Lock()
							// re-check after acquiring write lock to avoid double rotation
							if e.shouldCreateNewSegment(e.SegmentsCount) {
								if err := e.rotateSegment(); err != nil {
									lg.Printf("Error rotating segment: %v\n", err)
								}
							}
							e.mu.Unlock()
						}
						})
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					log.Println(err.Error())
				}
			}
		}()

		if err := watcher.Add(e.Dir); err != nil {
			lg.Fatalln("Failed to add path", err.Error())
		}
		<-make(chan struct{})
	}()

	// Sync ticker — batches fsyncs instead of one per write
	e.startSyncTicker()

	// Worker goroutines
	for range e.numWorkers {
		go func() {
			for work := range e.WorkersChan {
				if err := e.StoreLog(work.log); err != nil {
					work.errChan <- err
				}
				work.resChan <- "Log Stored Successfully"
			}
		}()
	}
}

func (e *Engine) GetFileName(segment int) string {
	return e.Dir + "/segment_" + strconv.Itoa(segment) + ".log"
}

func (e *Engine) shouldCreateNewSegment(currSegment int) bool {
	fileName := e.GetFileName(currSegment)
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return false
	}
	return fileInfo.Size() >= e.SegmentSize
}

func (e *Engine) StoreLog(log string) error {
	
	_, err := e.currentWriter.WriteString(log + "\n")
	return err
}
