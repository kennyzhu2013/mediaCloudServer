package clock

import (
	"errors"
	"sync"
	"time"
)

// TODO:先暂时不用精细化.
type Clock struct {
	lastNow int64

	// lastSystemTime is the last system time in the previous Update.
	// lastSystemTime indicates the logical time in the game, so this can be bigger than the current time.
	lastSystemTime int64

	currentFPS  float64
	currentTPS  float64
	lastUpdated int64
	fpsCount    int
	initTime    time.Time
	tpsCount    int

	m sync.Mutex
}

func (c *Clock) now() int64 {
	// time.Since() returns monotonic timer difference (#875):
	// https://golang.org/pkg/time/#hdr-Monotonic_Clocks
	return int64(time.Since(c.initTime))
}

func NewClock() *Clock {
	c := &Clock{fpsCount: 0, tpsCount: 0, initTime: time.Now()}
	n := c.now()
	c.lastNow = n
	c.lastSystemTime = n
	c.lastUpdated = n
	return c
}

func (c *Clock) CurrentFPS() float64 {
	c.m.Lock()
	v := c.currentFPS
	c.m.Unlock()
	return v
}

func (c *Clock) CurrentTPS() float64 {
	c.m.Lock()
	v := c.currentTPS
	c.m.Unlock()
	return v
}

func max(a, b int64) int64 {
	if a < b {
		return b
	}
	return a
}

func (c *Clock) calcCountFromTPS(tps int64, now int64) int {
	if tps == 0 {
		return 0
	}
	if tps < 0 {
		panic("clock: tps must >= 0")
	}

	diff := now - c.lastSystemTime
	if diff < 0 {
		return 0
	}

	count := 0
	syncWithSystemClock := false

	// Detect whether the previous time is too old.
	// Use either 5 ticks or 5/60 sec in the case when TPS is too big like 300 (#1444).
	if diff > max(int64(time.Second)*5/tps, int64(time.Second)*5/60) {
		// The previous time is too old.
		// Let's force to sync the game time with the system clock.
		syncWithSystemClock = true
	} else {
		count = int(diff * tps / int64(time.Second))
	}

	// Stabilize the count.
	// Without this adjustment, count can be unstable like 0, 2, 0, 2, ...
	// TODO: Brush up this logic so that this will work with any FPS. Now this works only when FPS = TPS.
	if count == 0 && (int64(time.Second)/tps/2) < diff {
		count = 1
	}
	if count == 2 && (int64(time.Second)/tps*3/2) > diff {
		count = 1
	}

	if syncWithSystemClock {
		c.lastSystemTime = now
	} else {
		c.lastSystemTime += int64(count) * int64(time.Second) / tps
	}

	return count
}

func (c *Clock) updateFPSAndTPS(now int64, count int) error {
	if now < c.lastUpdated {
		return errors.New("clock: lastUpdated must be older than now")
	}
	if time.Second > time.Duration(now-c.lastUpdated) {
		return nil
	}
	c.fpsCount++
	c.tpsCount += count
	c.currentFPS = float64(c.fpsCount) * float64(time.Second) / float64(now-c.lastUpdated)
	c.currentTPS = float64(c.tpsCount) * float64(time.Second) / float64(now-c.lastUpdated)
	c.lastUpdated = now
	c.fpsCount = 0
	c.tpsCount = 0
	return nil
}

const SyncWithFPS = -1

// Update updates the inner clock state and returns an integer value
// indicating how many times the game should update based on given tps.
// tps represents TPS (ticks per second).
// If tps is SyncWithFPS, Update always returns 1.
// If tps <= 0 and not SyncWithFPS, Update always returns 0.
//
// Update is expected to be called per frame.
func (c *Clock) Update(tps int) (int, error) {
	c.m.Lock()
	defer c.m.Unlock()

	n := c.now()
	if c.lastNow > n {
		// This ensures that now() must be monotonic (#875).
		return 0, errors.New("clock: lastNow must be older than n")
	}
	c.lastNow = n

	count := 0
	if tps == SyncWithFPS {
		count = 1
	} else if tps > 0 {
		count = c.calcCountFromTPS(int64(tps), n)
	}
	err := c.updateFPSAndTPS(n, count)
	if err != nil {
		count = 0
	}

	return count, err
}
