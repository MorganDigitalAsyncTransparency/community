// Spec: specs/api/api-contract.md (AC-27)
// Tests: backend/api/contract_test.go
package api

import "net/http"

func (s *Server) handleConfig(w http.ResponseWriter, _ *http.Request) {
	type areaResp struct {
		Name       string `json:"name"`
		PrimaryTag string `json:"primaryTag"`
	}
	type sloResp struct {
		FirstReplyHours int `json:"firstReplyHours"`
		ResolutionHours int `json:"resolutionHours"`
		InactivityHours int `json:"inactivityHours"`
	}
	type tagResp struct {
		Area                 string  `json:"area"`
		AreaIsDefault        bool    `json:"areaIsDefault"`
		StalledDays          int     `json:"stalledDays"`
		StalledDaysIsDefault bool    `json:"stalledDaysIsDefault"`
		SLO                  sloResp `json:"slo"`
		SLOIsDefault         bool    `json:"sloIsDefault"`
		ClosedTag            *string `json:"closedTag"`
	}
	type defaultsResp struct {
		StalledDays int     `json:"stalledDays"`
		Area        string  `json:"area"`
		SLO         sloResp `json:"slo"`
	}

	areas := make([]areaResp, len(s.TagConfig.Areas))
	for i, a := range s.TagConfig.Areas {
		areas[i] = areaResp{Name: a.Name, PrimaryTag: a.PrimaryTag}
	}

	tags := make(map[string]tagResp, len(s.ResolvedTags))
	for name, rt := range s.ResolvedTags {
		tags[name] = tagResp{
			Area: rt.Area, AreaIsDefault: rt.AreaIsDefault,
			StalledDays: rt.StalledDays, StalledDaysIsDefault: rt.StalledDaysIsDefault,
			SLO: sloResp{
				FirstReplyHours: rt.SLO.FirstReplyHours,
				ResolutionHours: rt.SLO.ResolutionHours,
				InactivityHours: rt.SLO.InactivityHours,
			},
			SLOIsDefault: rt.SLOIsDefault,
			ClosedTag:    rt.ClosedTag,
		}
	}

	defaults := defaultsResp{
		StalledDays: s.TagConfig.Defaults.StalledDays,
		Area:        s.TagConfig.Defaults.Area,
		SLO: sloResp{
			FirstReplyHours: s.TagConfig.Defaults.SLO.FirstReplyHours,
			ResolutionHours: s.TagConfig.Defaults.SLO.ResolutionHours,
			InactivityHours: s.TagConfig.Defaults.SLO.InactivityHours,
		},
	}

	respondJSON(w, map[string]any{
		"areas":                      areas,
		"tags":                       tags,
		"defaults":                   defaults,
		"distributionBucketCeilings": s.BucketCeilings,
	})
}
