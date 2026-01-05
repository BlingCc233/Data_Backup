package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/robfig/cron/v3"
)

type TaskType string

const (
	TaskTypeSchedule TaskType = "schedule"
	TaskTypeWatch    TaskType = "watch"
)

type TaskConfig struct {
	SourcePaths     []string     `json:"sourcePaths"`
	DestinationDir  string       `json:"destinationDir"`
	Filters         FilterConfig  `json:"filters"`
	UseCompression  bool         `json:"useCompression"`
	UseEncryption   bool         `json:"useEncryption"`
	Algorithm       uint8        `json:"algorithm"`
	Password        string       `json:"password"`
	Incremental     bool         `json:"incremental"`
	WatchDebounceMs int          `json:"watchDebounceMs"`
	CronExpr        string       `json:"cronExpr"`
	WatchPaths      []string     `json:"watchPaths"`
	CreatedAt       time.Time    `json:"createdAt"`
	UpdatedAt       time.Time    `json:"updatedAt"`
	LastBackupPath  string       `json:"lastBackupPath"`
}

type BackupTask struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Type    TaskType `json:"type"`
	Enabled bool     `json:"enabled"`
	Config  TaskConfig `json:"config"`
}

type TaskExecutor func(ctx context.Context, task BackupTask) (newBackupPath string, err error)

type TaskRunner struct {
	mu       sync.Mutex
	tasks    map[string]*taskState
	executor TaskExecutor

	cron    *cron.Cron
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
}

type taskState struct {
	task BackupTask

	cronEntry cron.EntryID

	watcher   *fsnotify.Watcher
	watchDone chan struct{}
	debounce  *time.Timer

	running bool
	pending bool
}

func NewTaskRunner(executor TaskExecutor) *TaskRunner {
	return &TaskRunner{
		tasks:    make(map[string]*taskState),
		executor: executor,
		cron:     cron.New(),
	}
}

func (tr *TaskRunner) Start() {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	if tr.started {
		return
	}
	tr.ctx, tr.cancel = context.WithCancel(context.Background())
	tr.started = true
	tr.cron.Start()

	for id := range tr.tasks {
		_ = tr.applyTaskLocked(id)
	}
}

func (tr *TaskRunner) Stop() {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	if !tr.started {
		return
	}

	if tr.cancel != nil {
		tr.cancel()
	}
	tr.cron.Stop()

	for id := range tr.tasks {
		tr.stopTaskLocked(id)
	}
	tr.started = false
}

func (tr *TaskRunner) Upsert(task BackupTask) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	st, ok := tr.tasks[task.ID]
	if !ok {
		st = &taskState{task: task}
		tr.tasks[task.ID] = st
	} else {
		st.task = task
	}

	if tr.started {
		return tr.applyTaskLocked(task.ID)
	}
	return nil
}

func (tr *TaskRunner) Remove(taskID string) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.stopTaskLocked(taskID)
	delete(tr.tasks, taskID)
}

func (tr *TaskRunner) RunNow(taskID string) {
	tr.runTask(taskID)
}

func (tr *TaskRunner) List() []BackupTask {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	out := make([]BackupTask, 0, len(tr.tasks))
	for _, st := range tr.tasks {
		out = append(out, st.task)
	}
	return out
}

func (tr *TaskRunner) applyTaskLocked(taskID string) error {
	st, ok := tr.tasks[taskID]
	if !ok {
		return nil
	}

	// Clear previous state first.
	tr.stopTaskLocked(taskID)

	if !st.task.Enabled {
		return nil
	}

	switch st.task.Type {
	case TaskTypeSchedule:
		entryID, err := tr.cron.AddFunc(st.task.Config.CronExpr, func() {
			tr.runTask(taskID)
		})
		if err != nil {
			return err
		}
		st.cronEntry = entryID
	case TaskTypeWatch:
		if err := tr.startWatchLocked(taskID); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported task type: %s", st.task.Type)
	}

	return nil
}

func (tr *TaskRunner) stopTaskLocked(taskID string) {
	st, ok := tr.tasks[taskID]
	if !ok {
		return
	}

	if st.cronEntry != 0 {
		tr.cron.Remove(st.cronEntry)
		st.cronEntry = 0
	}

	if st.debounce != nil {
		st.debounce.Stop()
		st.debounce = nil
	}

	if st.watcher != nil {
		close(st.watchDone)
		_ = st.watcher.Close()
		st.watcher = nil
	}
}

func (tr *TaskRunner) startWatchLocked(taskID string) error {
	st, ok := tr.tasks[taskID]
	if !ok {
		return nil
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	for _, p := range st.task.Config.WatchPaths {
		if err := addWatchRecursive(watcher, p); err != nil {
			_ = watcher.Close()
			return err
		}
	}

	st.watcher = watcher
	st.watchDone = make(chan struct{})

	debounce := time.Duration(st.task.Config.WatchDebounceMs) * time.Millisecond
	if debounce <= 0 {
		debounce = 500 * time.Millisecond
	}

	go func() {
		for {
			select {
			case <-st.watchDone:
				return
			case <-tr.ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// If a new directory appears, add it to watcher (best-effort).
				if event.Op&fsnotify.Create != 0 {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						_ = addWatchRecursive(watcher, event.Name)
					}
				}

				tr.requestRun(taskID, debounce)
			case <-watcher.Errors:
				// Ignore watcher errors; tasks can still be run manually.
			}
		}
	}()

	return nil
}

func addWatchRecursive(w *fsnotify.Watcher, root string) error {
	info, err := os.Stat(root)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		// Watch the parent directory for file changes.
		return w.Add(filepath.Dir(root))
	}

	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return w.Add(path)
		}
		return nil
	})
}

func (tr *TaskRunner) requestRun(taskID string, debounce time.Duration) {
	tr.mu.Lock()
	st, ok := tr.tasks[taskID]
	if !ok || !st.task.Enabled {
		tr.mu.Unlock()
		return
	}

	if st.debounce != nil {
		st.debounce.Stop()
	}
	st.debounce = time.AfterFunc(debounce, func() {
		tr.runTask(taskID)
	})
	tr.mu.Unlock()
}

func (tr *TaskRunner) runTask(taskID string) {
	tr.mu.Lock()
	st, ok := tr.tasks[taskID]
	if !ok || !st.task.Enabled {
		tr.mu.Unlock()
		return
	}
	if st.running {
		st.pending = true
		tr.mu.Unlock()
		return
	}
	st.running = true
	taskCopy := st.task
	tr.mu.Unlock()

	ctx := tr.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	newPath, err := tr.executor(ctx, taskCopy)

	tr.mu.Lock()
	defer tr.mu.Unlock()
	st.running = false

	if err == nil && newPath != "" {
		st.task.Config.LastBackupPath = newPath
	}

	if st.pending {
		st.pending = false
		go tr.runTask(taskID)
	}
}

