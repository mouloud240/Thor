// Storage Engine
// Architecture: single writer goroutine owns all file operations.
// Watcher and sync ticker send commands through writerChannel — they never
// touch the file system directly. This eliminates lock contention and
// data races on currentWriter / currentFile.
package storage

import (
	"bufio"
	"fmt"
	"io"
	lg "log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// WorkerInput carries a log entry and its response channels through the pipeline.
type WorkerInput struct {
	errChan chan<- error
	log     string
	resChan chan<- string
}

func NewWorkerInput(errChan chan<- error, log string, resChan chan<- string) WorkerInput {
	return WorkerInput{errChan: errChan, log: log, resChan: resChan}
}

// cmdKind identifies what the writer goroutine should do.
type cmdKind int

const (
	cmdWrite  cmdKind = iota // write a log entry
	cmdFlush                 // flush + fsync without rotating
	cmdRotate                // flush + fsync + rotate to new segment
)

// writerCmd is the only way anything communicates with the writer goroutine.
// For cmdRotate, done is closed once rotation completes so the watcher can
// unblock. For cmdWrite and cmdFlush, done may be nil.
type writerCmd struct {
	kind cmdKind
	log  string
	done chan struct{}
}

// Engine is the storage engine. After construction, StartWorkers must be
// called before sending any WorkerInput.
type Engine struct {
	Dir           string
	SegmentSize   int64
	SegmentsCount int
	numWorkers    int

	// writerChannel is the single entry point for all file operations.
	// Only the writer goroutine reads from it.
	writerChannel chan writerCmd

	// WorkersChan receives work from the pipeline layer.
	WorkersChan chan WorkerInput

	// mu guards SegmentsCount for reads from the watcher goroutine.
	// It is never held during file I/O.
	mu sync.RWMutex

	// These fields are owned exclusively by the writer goroutine.
	// No other goroutine may read or write them.
	currentFile   *os.File
	currentWriter *bufio.Writer
}

func NewEngine(dir string, segmentSize int64, numWorkers int) (*Engine, error) {
	lg.Printf("Initializing storage engine — dir: %s  segment size: %d  workers: %d\n",
		dir, segmentSize, numWorkers)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Find the segment to resume writing into.
	segmentCount := 0
	for {
		fileName := dir + "/segment_" + strconv.Itoa(segmentCount) + ".log"
		stat, err := os.Stat(fileName)
		if os.IsNotExist(err) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("stat segment %d: %w", segmentCount, err)
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
		WorkersChan:   make(chan WorkerInput, numWorkers),
		writerChannel: make(chan writerCmd, numWorkers*50),
	}

	if err := e.openCurrentSegment(); err != nil {
		return nil, err
	}

	return e, nil
}

// openCurrentSegment opens the active segment file and sets up the buffered
// writer. Must only be called from the writer goroutine (or during init,
// before StartWorkers).
func (e *Engine) openCurrentSegment() error {
	fileName := e.GetFileName(e.SegmentsCount)
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open segment %s: %w", fileName, err)
	}
	e.currentFile = file
	e.currentWriter = bufio.NewWriterSize(file, 256*1024) // 256 KB buffer
	return nil
}

// rotateSegment flushes and closes the current segment, then opens a new one.
// Must only be called from the writer goroutine.
func (e *Engine) rotateSegment() error {
	if err := e.currentWriter.Flush(); err != nil {
		return fmt.Errorf("flush before rotation: %w", err)
	}
	if err := e.currentFile.Sync(); err != nil {
		return fmt.Errorf("sync before rotation: %w", err)
	}
	if err := e.currentFile.Close(); err != nil {
		return fmt.Errorf("close before rotation: %w", err)
	}

	e.mu.Lock()
	e.SegmentsCount++
	e.mu.Unlock()

	lg.Printf("Rotating to new segment: %s\n", e.GetFileName(e.SegmentsCount))
	return e.openCurrentSegment()
}

// writeFullString writes all bytes of s into w, retrying on short writes.
// Must only be called from the writer goroutine.
func writeFullString(w *bufio.Writer, s string) error {
	remaining := s
	for len(remaining) > 0 {
		n, err := w.WriteString(remaining)
		remaining = remaining[n:]
		if err != nil && err != io.ErrShortWrite {
			return fmt.Errorf("writeFullString: %w", err)
		}
	}
	return nil
}

// StartWorkers launches all background goroutines. Call once after NewEngine.
func (e *Engine) StartWorkers() {
	// ── Writer goroutine ──────────────────────────────────────────────────────
	// The only goroutine that may touch currentFile or currentWriter.
	// All other goroutines communicate via writerChannel.
	go func() {
		for cmd := range e.writerChannel {
			switch cmd.kind {

			case cmdWrite:
				entry := cmd.log
				// Ensure every entry ends with a newline.
				if len(entry) == 0 || entry[len(entry)-1] != '\n' {
					entry += "\n"
				}
				if err := writeFullString(e.currentWriter, entry); err != nil {
					lg.Printf("Write error: %v\n", err)
				}

			case cmdFlush:
				if err := e.currentWriter.Flush(); err != nil {
					lg.Printf("Flush error: %v\n", err)
				}
				if err := e.currentFile.Sync(); err != nil {
					lg.Printf("Sync error: %v\n", err)
				}

			case cmdRotate:
				// Flush and sync are implicit in rotateSegment.
				if err := e.rotateSegment(); err != nil {
					lg.Printf("Rotation error: %v\n", err)
				}
				// Signal the watcher that rotation is complete.
				if cmd.done != nil {
					close(cmd.done)
				}
			}
		}
	}()

	// ── Watcher goroutine ─────────────────────────────────────────────────────
	// Observes file-system events and signals the writer to rotate.
	// Never touches the file system for writes itself.
	go func() {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			lg.Fatalf("Failed to start watcher: %v\n", err)
		}

		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					if !event.Has(fsnotify.Write) {
						continue
					}

					e.mu.RLock()
					count := e.SegmentsCount
					e.mu.RUnlock()

					if !e.shouldCreateNewSegment(count) {
						continue
					}

					// Signal the writer to rotate and wait for it to finish
					// before checking the segment again. This prevents the
					// watcher from flooding writerChannel with rotate commands.
					done := make(chan struct{})
					e.writerChannel <- writerCmd{kind: cmdRotate, done: done}
					<-done

				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					lg.Printf("Watcher error: %v\n", err)
				}
			}
		}()

		if err := watcher.Add(e.Dir); err != nil {
			lg.Fatalf("Failed to watch dir %s: %v\n", e.Dir, err)
		}
		// Keep the goroutine alive.
		<-make(chan struct{})
	}()

	// ── Sync ticker ───────────────────────────────────────────────────────────
	// Batches fsyncs to avoid one syscall per write. Sends a flush command
	// into writerChannel — never calls Flush or Sync directly.
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			e.writerChannel <- writerCmd{kind: cmdFlush}
		}
	}()

	// ── Pipeline worker goroutines ────────────────────────────────────────────
	// Receive parsed log entries from the pipeline layer and forward them
	// to the writer via StoreLog.
	for range e.numWorkers {
		go func() {
			for work := range e.WorkersChan {
				if err := e.StoreLog(work.log); err != nil {
					work.errChan <- err
					continue
				}
				work.resChan <- "Log stored successfully"
			}
		}()
	}
}

// StoreLog enqueues a log entry for the writer goroutine. It is safe to call
// from any goroutine.
func (e *Engine) StoreLog(log string) error {
	e.writerChannel <- writerCmd{kind: cmdWrite, log: log}
	return nil
}

func (e *Engine) GetFileName(segment int) string {
	return e.Dir + "/segment_" + strconv.Itoa(segment) + ".log"
}

func (e *Engine) shouldCreateNewSegment(currSegment int) bool {
	info, err := os.Stat(e.GetFileName(currSegment))
	if err != nil {
		return false
	}
	return info.Size() >= e.SegmentSize
}
