package server

import (
	"testing"

	"github.com/liangzh77/keychain/internal/admin"
)

func TestBuildHistoryChartPlacesSparsePointsComfortably(t *testing.T) {
	_, _, onePoint, _ := buildHistoryChart([]admin.DispatchHistoryPoint{
		{BucketStart: "2026-05-09T00:00:00Z", TotalCount: 4, FailedCount: 1},
	})
	if len(onePoint) != 1 || onePoint[0].X != 500 {
		t.Fatalf("one point x = %#v, want centered at 500", onePoint)
	}

	_, _, twoPoints, _ := buildHistoryChart([]admin.DispatchHistoryPoint{
		{BucketStart: "2026-05-08T00:00:00Z", TotalCount: 2, FailedCount: 0},
		{BucketStart: "2026-05-09T00:00:00Z", TotalCount: 4, FailedCount: 1},
	})
	if len(twoPoints) != 2 || twoPoints[0].X != 357 || twoPoints[1].X != 643 {
		t.Fatalf("two point x = %#v, want one-third and two-thirds", twoPoints)
	}
}
