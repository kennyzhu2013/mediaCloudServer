package hooks


import (
	"sync"
)

var m sync.Mutex

var onBeforeUpdateHooks = []func() error{}

// AppendHookOnBeforeUpdate appends a hook function that is run before the main update function
// every frame.
func AppendHookOnBeforeUpdate(f func() error) {
	m.Lock()
	onBeforeUpdateHooks = append(onBeforeUpdateHooks, f)
	m.Unlock()
}

func RunBeforeUpdateHooks() error {
	m.Lock()
	defer m.Unlock()

	for _, f := range onBeforeUpdateHooks {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

var (
	audioSuspended bool
	onSuspendAudio func() error
	onResumeAudio  func() error
)

func OnSuspendAudio(f func() error) {
	m.Lock()
	onSuspendAudio = f
	m.Unlock()
}

func OnResumeAudio(f func() error) {
	m.Lock()
	onResumeAudio = f
	m.Unlock()
}

func SuspendAudio() error {
	m.Lock()
	defer m.Unlock()
	if audioSuspended {
		return nil
	}
	audioSuspended = true
	if onSuspendAudio != nil {
		return onSuspendAudio()
	}
	return nil
}

func ResumeAudio() error {
	m.Lock()
	defer m.Unlock()
	if !audioSuspended {
		return nil
	}
	audioSuspended = false
	if onResumeAudio != nil {
		return onResumeAudio()
	}
	return nil
}
