package task

import "sync/atomic"

type StatisticsInfo struct {
	TotalCount int64

	//successCount int
	successCount *atomic.Int64

	//failedCount int
	failedCount *atomic.Int64

	//diffCount int
	diffCount *atomic.Int64
}

func NewStatisticsInfo(totalCount int) *StatisticsInfo {

	s := &StatisticsInfo{
		TotalCount:   int64(totalCount),
		successCount: &atomic.Int64{},
		failedCount:  &atomic.Int64{},
		diffCount:    &atomic.Int64{},
	}

	s.successCount.Store(0)
	s.failedCount.Store(0)
	s.diffCount.Store(0)

	return s
}

func (s *StatisticsInfo) AddSuccess() {
	s.successCount.Add(1)
}

func (s *StatisticsInfo) AddFailed() {
	s.failedCount.Add(1)
}

func (s *StatisticsInfo) AddDiff() {
	s.diffCount.Add(1)
}

func (s *StatisticsInfo) GetTotalCount() int64 {
	return s.TotalCount
}

func (s *StatisticsInfo) GetSuccessCount() int64 {
	return s.successCount.Load()
}

func (s *StatisticsInfo) GetFailedCount() int64 {
	return s.failedCount.Load()
}

func (s *StatisticsInfo) GetDiffCount() int64 {
	return s.diffCount.Load()
}
