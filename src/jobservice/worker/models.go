package worker

// Stats represents the healthy and status of all the running worker pools.
type Stats struct {
	Pools []*StatsData `json:"worker_pools"`
}

// StatsData represents the healthy and status of the worker worker.
type StatsData struct {
	WorkerPoolID string   `json:"worker_pool_id"`
	StartedAt    int64    `json:"started_at"`
	HeartbeatAt  int64    `json:"heartbeat_at"`
	JobNames     []string `json:"job_names"`
	Concurrency  uint     `json:"concurrency"`
	Status       string   `json:"status"`
}
