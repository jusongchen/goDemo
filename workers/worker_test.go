package workers

import (
	"errors"
	"math"
	"testing"
	"time"
)

var WorkerTests = []struct {
	// c           *Context
	numTask     int
	DOP         int
	expectedSec float64
	expectedErr error
}{
	{1, 1, 1, nil},
	{1, 2, 1, nil},
	{2, 1, 2, nil},
	{2, 2, 1, nil},
	{2, 3, 1, nil},
	{3, 1, 3, nil},
	{3, 2, 2, nil},
	{3, 3, 1, nil},
	{3, 4, 1, nil},
}

type timer struct {
	id    int
	start time.Time
	end   time.Time
}

func (tm *timer) Exec(id WorkerID) error {
	tm.start = time.Now()
	time.Sleep(time.Second)
	tm.end = time.Now()
	// log.Printf("worker ID %d timer %v %v\n", id, tm.start, tm.end)
	return nil
}

func factoryFuncNoErr(timers []*timer) FactoryFunc {
	var index int
	return func() Task {
		if index == len(timers) { //
			return nil
		}
		tm := timers[index]
		index++
		return tm
	}
}

func createTimers(numTimer int) []*timer {
	timers := []*timer{}
	for i := 0; i < numTimer; i++ {
		timers = append(timers, &timer{id: i})
	}
	return timers
}

var errZeroTime = errors.New("start or end time not set")

func calcuateElapsedTime(timers []*timer) (time.Duration, int) {
	minStart := time.Now()
	maxEnd := time.Now().Add(-time.Hour)
	numNotExecuted := 0

	for _, tm := range timers {
		if (tm.start == time.Time{}) || (tm.end == time.Time{}) {
			numNotExecuted++
		}
		if tm.start.Before(minStart) {
			minStart = tm.start
		}
		if tm.end.After(maxEnd) {
			maxEnd = tm.end
		}
	}
	return maxEnd.Sub(minStart), numNotExecuted
}

//TestDo test Do func which execute timers in parallel
func TestDo(t *testing.T) {
	for _, tt := range WorkerTests {

		timers := createTimers(tt.numTask)

		ctx := Context{
			DOP:         tt.DOP,
			FactoryFunc: factoryFuncNoErr(timers),
		}

		err := Do(&ctx)
		if err != tt.expectedErr {
			t.Errorf("expected err %v, actual err %v", tt.expectedErr, err)
		}
		actualDuration, numNotExecuted := calcuateElapsedTime(timers)
		if numNotExecuted != 0 {
			t.Errorf("Do(%v): expected %d tasks be executed, actual %d tasks not executed", ctx, tt.numTask, numNotExecuted)
		}

		actualSec := actualDuration.Seconds()
		if math.Abs(actualSec-tt.expectedSec) >= 1 {
			t.Errorf("Do(%v): expected tasks to complete in %v, actual %v", ctx, tt.expectedSec, actualSec)
		}
	}
}
