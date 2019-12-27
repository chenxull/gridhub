package mgt

import (
	"context"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/common/query"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/common/utils"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/errs"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/job"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
)

// Manager  defies the related operations to handle the management of job stats
type Manager interface {

	//Get the stats data of all kinds of jobs
	GetJobs(q *query.Parameter) ([]*job.Stats, int64, error)

	// Get the executions of the specified periodic job by pagination
	//
	// Arguments:
	//   pID: ID of the periodic job
	//   q *query.Parameter: query parameters
	//
	// Returns:
	//   The matched job stats list,
	//   The total number of the executions,
	//   Non nil error if any issues meet.
	GetPeriodicExecution(pID string, q *query.Parameter) ([]*job.Stats, int64, error)
	// Get the scheduled jobs
	//
	// Arguments:
	//   q *query.Parameter: query parameters
	//
	// Returns:
	//   The matched job stats list,
	//   The total number of the executions,
	//   Non nil error if any issues meet.
	GetScheduledJobs(q *query.Parameter) ([]*job.Stats, int64, error)

	// Get the stats of the specified job
	//
	// Arguments:
	//   jobID string: ID of the job
	//
	// Returns:
	//   The job stats
	//   Non nil error if any issues meet
	GetJob(jobID string) (*job.Stats, error)

	// Save the job stats
	//
	// Arguments:
	//   job *job.Stats: the saving job stats
	//
	// Returns:
	//   Non nil error if any issues meet
	SaveJob(job *job.Stats) error
}

// basicManager is the default implementation of @manager,
type basicManager struct {
	//system context  稍微复杂一点的结构体都需要context 来携带上下文信息
	ctx context.Context
	// db namespace redis
	namespace string
	// redis conn pool
	pool *redis.Pool
	// go work client,用来控制 job 执行的框架，需要 redis 配合
	client *work.Client
}

// NewManager news a basic manager
func NewManager(ctx context.Context, ns string, pool *redis.Pool) Manager {
	return &basicManager{
		ctx:       ctx,
		namespace: ns,
		pool:      pool,
		client:    work.NewClient(ns, pool),
	}
}

// GetJobs is implementation of Manager.GetJobs
// Because of the hash set used to keep the job stats, we can not support a standard pagination.
// A cursor is used to fetch the jobs with several batches.
func (bm *basicManager) GetJobs(q *query.Parameter) ([]*job.Stats, int64, error) {
	//todo
	panic("implement me")
}

func (bm *basicManager) GetPeriodicExecution(pID string, q *query.Parameter) ([]*job.Stats, int64, error) {
	//todo
	panic("implement me")
}

func (bm *basicManager) GetScheduledJobs(q *query.Parameter) ([]*job.Stats, int64, error) {
	//todo
	panic("implement me")
}

func (bm *basicManager) GetJob(jobID string) (*job.Stats, error) {
	if utils.IsEmptyStr(jobID) {
		return nil, errs.BadRequestError("empty job ID")
	}

	// 从job tracker 中获取job的准确信息。其实也是通过 Load 方法实时在redis 中获取
	t := job.NewBasicTrackerWithID(bm.ctx, jobID, bm.namespace, bm.pool, nil)
	if err := t.Load(); err != nil {
		return nil, err
	}
	return t.Job(), nil
}

func (bm *basicManager) SaveJob(j *job.Stats) error {
	if j == nil {
		return errs.BadRequestError("nil saving job stats")
	}

	t := job.NewBasicTrackerWithStats(bm.ctx, j, bm.namespace, bm.pool, nil)
	return t.Save()
}

// queryExecutions queries periodic executions by status
func queryExecutions(conn redis.Conn, dataKey string, q *query.Parameter) ([]string, int64, error) {
	total, err := redis.Int64(conn.Do("ZCOUNT", dataKey, 0, "+inf"))
	if err != nil {
		return nil, 0, err
	}

	var pageNumber, pageSize uint = 1, query.DefaultPageSize
	if q.PageNumber > 0 {
		pageNumber = q.PageNumber
	}
	if q.PageSize > 0 {
		pageSize = q.PageSize
	}

	results := make([]string, 0)
	if total == 0 || (int64)((pageNumber-1)*pageSize) >= total {
		return results, total, nil
	}

	offset := (pageNumber - 1) * pageSize
	args := []interface{}{dataKey, "+inf", 0, "LIMIT", offset, pageSize}

	eIDs, err := redis.Values(conn.Do("ZREVRANGEBYSCORE", args...))
	if err != nil {
		return nil, 0, err
	}

	for _, eID := range eIDs {
		if eIDBytes, ok := eID.([]byte); ok {
			results = append(results, string(eIDBytes))
		}
	}

	return results, total, nil
}
