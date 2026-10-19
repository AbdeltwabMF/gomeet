// Package schedule provides interval matching and occurrence deduplication
// shared by the local and Google calendar monitors.
package schedule

import (
	"sync"
	"time"
)

// Occurrence reports whether a scheduled start falls inside the interval
// (after, through]. A zero after means the interval starts at the beginning
// of time, which is used for the initial catch-up window.
func Occurrence(start, after, through time.Time) bool {
	if start.After(through) {
		return false
	}
	if after.IsZero() {
		return true
	}

	return start.After(after)
}

// Deduper suppresses repeated actions for the same occurrence. It is safe for
// concurrent use.
type Deduper struct {
	mu    sync.Mutex
	seen  map[string]time.Time
	limit int
}

// NewDeduper returns a Deduper that remembers at most limit occurrences.
// Older entries are pruned as new ones are recorded.
func NewDeduper(limit int) *Deduper {
	if limit <= 0 {
		limit = 512
	}

	return &Deduper{seen: make(map[string]time.Time), limit: limit}
}

// First reports whether the key is new. It returns true and records the key
// when the occurrence has not been seen before.
func (d *Deduper) First(key string, occurrence, now time.Time) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.seen[key]; ok {
		return false
	}

	if len(d.seen) >= d.limit {
		d.prune(now)
	}
	d.seen[key] = occurrence

	return true
}

// prune drops entries that are older than the retention window.
func (d *Deduper) prune(now time.Time) {
	const retention = 24 * time.Hour

	for key, occurrence := range d.seen {
		if now.Sub(occurrence) > retention {
			delete(d.seen, key)
		}
	}

	// The window may not be enough to free space, so drop the oldest entries.
	for len(d.seen) >= d.limit {
		oldestKey := ""
		var oldest time.Time

		for key, occurrence := range d.seen {
			if oldestKey == "" || occurrence.Before(oldest) {
				oldestKey, oldest = key, occurrence
			}
		}
		if oldestKey == "" {
			return
		}
		delete(d.seen, oldestKey)
	}
}
