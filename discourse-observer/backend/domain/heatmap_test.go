package domain

import (
	"testing"
	"time"

	"github.com/code-community/discourse-observer/backend/model"
)

func TestBuildHeatmap(t *testing.T) {
	// Monday 2026-03-16 10:30 UTC → day 0, hour 10
	topics := []model.Topic{
		{CreatedAt: time.Date(2026, 3, 16, 10, 30, 0, 0, time.UTC)},
	}
	data := BuildHeatmap(topics)

	if data.Cells[0][10].Count != 1 {
		t.Errorf("Monday 10:30 cell: got %d, want 1", data.Cells[0][10].Count)
	}
	if data.MaxCount != 1 {
		t.Errorf("maxCount: got %d, want 1", data.MaxCount)
	}
}

func TestBuildHeatmapMultipleSameSlot(t *testing.T) {
	monday10 := time.Date(2026, 3, 16, 10, 0, 0, 0, time.UTC)
	topics := []model.Topic{
		{CreatedAt: monday10},
		{CreatedAt: monday10.Add(15 * time.Minute)},
		{CreatedAt: monday10.Add(45 * time.Minute)},
	}
	data := BuildHeatmap(topics)

	if data.Cells[0][10].Count != 3 {
		t.Errorf("3 topics same slot: got %d, want 3", data.Cells[0][10].Count)
	}
	if data.MaxCount != 3 {
		t.Errorf("maxCount: got %d, want 3", data.MaxCount)
	}
}

func TestBuildHeatmapEmpty(t *testing.T) {
	data := BuildHeatmap(nil)
	if data.MaxCount != 0 {
		t.Errorf("empty maxCount: got %d, want 0", data.MaxCount)
	}
	for day := 0; day < 7; day++ {
		for hour := 0; hour < 24; hour++ {
			if data.Cells[day][hour].Count != 0 {
				t.Errorf("cell[%d][%d] not zero", day, hour)
			}
			if data.Cells[day][hour].Day != day || data.Cells[day][hour].Hour != hour {
				t.Errorf("cell[%d][%d] has wrong day/hour", day, hour)
			}
		}
	}
}

func TestBuildHeatmapSunday(t *testing.T) {
	// Sunday 2026-03-15 14:00 UTC → day 6
	topics := []model.Topic{
		{CreatedAt: time.Date(2026, 3, 15, 14, 0, 0, 0, time.UTC)},
	}
	data := BuildHeatmap(topics)
	if data.Cells[6][14].Count != 1 {
		t.Errorf("Sunday maps to day 6: got count %d", data.Cells[6][14].Count)
	}
}
