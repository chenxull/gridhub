package rds

import (
	"fmt"
	"strings"
)

// Functions defined here are mainly from dep lib "github.com/gocraft/work".
// Only for compatible

// 定义好 redis中各个字段
// RedisNamespacePrefix ... Same with 'KeyNamespacePrefix', only for compatibility.
func RedisNamespacePrefix(namespace string) string {
	return KeyNamespacePrefix(namespace)
}

// RedisKeyScheduled returns key of scheduled job.
func RedisKeyScheduled(namespace string) string {
	return RedisNamespacePrefix(namespace) + "scheduled"
}

// RedisKeyLastPeriodicEnqueue returns key of timestamp if last periodic enqueue.
func RedisKeyLastPeriodicEnqueue(namespace string) string {
	return RedisNamespacePrefix(namespace) + "last_periodic_enqueue_h"
}

// KeyNamespacePrefix returns the based key based on the namespace.
func KeyNamespacePrefix(namespace string) string {
	ns := strings.TrimSpace(namespace)
	if !strings.HasSuffix(ns, ":") {
		return fmt.Sprintf("%s:", ns)
	}

	return ns
}

// KeyPeriod returns the key of period
func KeyPeriod(namespace string) string {
	return fmt.Sprintf("%s%s", KeyNamespacePrefix(namespace), "period")
}

// KeyPeriodicPolicy returns the key of periodic policies.
func KeyPeriodicPolicy(namespace string) string {
	return fmt.Sprintf("%s:%s", KeyPeriod(namespace), "policies")
}

// KeyPeriodicNotification returns the key of periodic pub/sub channel.
func KeyPeriodicNotification(namespace string) string {
	return fmt.Sprintf("%s:%s", KeyPeriodicPolicy(namespace), "notifications")
}

// KeyPeriodicLock returns the key of locker under period
func KeyPeriodicLock(namespace string) string {
	return fmt.Sprintf("%s:%s", KeyPeriod(namespace), "lock")
}

// KeyJobStats returns the key of job stats
func KeyJobStats(namespace string, jobID string) string {
	return fmt.Sprintf("%s%s:%s", KeyNamespacePrefix(namespace), "job_stats", jobID)
}

// KeyUpstreamJobAndExecutions returns the key for persisting executions.
func KeyUpstreamJobAndExecutions(namespace, upstreamJobID string) string {
	return fmt.Sprintf("%s%s:%s", KeyNamespacePrefix(namespace), "executions", upstreamJobID)
}

// KeyHookEventRetryQueue returns the key of hook event retrying queue
func KeyHookEventRetryQueue(namespace string) string {
	return fmt.Sprintf("%s%s", KeyNamespacePrefix(namespace), "hook_events")
}

// KeyStatusUpdateRetryQueue returns the key of status change retrying queue
func KeyStatusUpdateRetryQueue(namespace string) string {
	return fmt.Sprintf("%s%s", KeyNamespacePrefix(namespace), "status_change_events")
}
