package server

import (
	"sync"

	"go.klarlabs.de/mcp/protocol"
)

// taskSubscriptions maps task IDs to the subscriptions/listen streams that
// asked for notifications/tasks (SEP-2663). Delivery uses the same
// ResourceNotifier the resource-subscription registry uses, keyed by
// subscription id (the SSE client id of the listen stream).
type taskSubscriptions struct {
	mu       sync.RWMutex
	byTask   map[string]map[string]struct{} // taskId -> set of subscription ids
	notifier ResourceNotifier
}

func newTaskSubscriptions() *taskSubscriptions {
	return &taskSubscriptions{byTask: make(map[string]map[string]struct{})}
}

func (r *taskSubscriptions) setNotifier(n ResourceNotifier) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.notifier = n
}

func (r *taskSubscriptions) subscribe(subID, taskID string) {
	if subID == "" || taskID == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	set := r.byTask[taskID]
	if set == nil {
		set = make(map[string]struct{})
		r.byTask[taskID] = set
	}
	set[subID] = struct{}{}
}

func (r *taskSubscriptions) removeClient(subID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, set := range r.byTask {
		delete(set, subID)
		if len(set) == 0 {
			delete(r.byTask, id)
		}
	}
}

func (r *taskSubscriptions) notify(t *AugTask) {
	if t == nil {
		return
	}
	r.mu.RLock()
	notifier := r.notifier
	set := r.byTask[t.TaskID]
	ids := make([]string, 0, len(set))
	for id := range set {
		ids = append(ids, id)
	}
	r.mu.RUnlock()
	if notifier == nil || len(ids) == 0 {
		return
	}
	params := TaskNotificationParams(t)
	for _, id := range ids {
		p := cloneTaskParams(params)
		meta, _ := p["_meta"].(map[string]any)
		if meta == nil {
			meta = map[string]any{}
		}
		meta[protocol.MetaKeySubscriptionID] = id
		p["_meta"] = meta
		_ = notifier.NotifyClient(id, protocol.MethodTasks, p)
	}
}

func cloneTaskParams(in map[string]any) map[string]any {
	out := make(map[string]any, len(in)+1)
	for k, v := range in {
		out[k] = v
	}
	return out
}
