// Spec: specs/api/api-contract.md (AC-26)
// Tests: backend/domain/heatmap_test.go
package domain

import "github.com/code-community/discourse-observer/backend/model"

// HeatmapCell holds the count for one day/hour combination.
type HeatmapCell struct {
	Day   int `json:"day"`
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

// HeatmapData holds the full 7×24 heatmap grid and peak count.
type HeatmapData struct {
	Cells    [7][24]HeatmapCell
	MaxCount int
}

// BuildHeatmap counts topic creation by UTC day-of-week and hour.
// Day 0 = Monday, Day 6 = Sunday.
func BuildHeatmap(topics []model.Topic) HeatmapData {
	var data HeatmapData

	for day := 0; day < 7; day++ {
		for hour := 0; hour < 24; hour++ {
			data.Cells[day][hour] = HeatmapCell{Day: day, Hour: hour}
		}
	}

	for i := range topics {
		utc := topics[i].CreatedAt.UTC()
		day := (int(utc.Weekday()) + 6) % 7
		hour := utc.Hour()
		data.Cells[day][hour].Count++
		if data.Cells[day][hour].Count > data.MaxCount {
			data.MaxCount = data.Cells[day][hour].Count
		}
	}

	return data
}
