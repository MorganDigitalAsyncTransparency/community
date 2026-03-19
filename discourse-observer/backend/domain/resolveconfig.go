// Spec: specs/api/api-contract.md (AC-27)
// Tests: backend/api/contract_test.go
package domain

import "github.com/code-community/discourse-observer/backend/model"

// ResolveAllTags resolves every tag in the config against defaults.
func ResolveAllTags(cfg *model.TagConfig) map[string]model.ResolvedTag {
	result := make(map[string]model.ResolvedTag, len(cfg.Tags))
	for name, spec := range cfg.Tags {
		result[name] = resolveTag(spec, cfg.Defaults)
	}
	return result
}

func resolveTag(spec model.TagSpec, defaults model.TagDefaults) model.ResolvedTag {
	rt := model.ResolvedTag{
		Area:                 defaults.Area,
		AreaIsDefault:        true,
		StalledDays:          defaults.StalledDays,
		StalledDaysIsDefault: true,
		SLO:                  defaults.SLO,
		SLOIsDefault:         true,
	}

	if spec.Area != "" {
		rt.Area = spec.Area
		rt.AreaIsDefault = false
	}
	if spec.StalledDays != nil {
		rt.StalledDays = *spec.StalledDays
		rt.StalledDaysIsDefault = false
	}
	if spec.SLO != nil {
		rt.SLO = *spec.SLO
		rt.SLOIsDefault = false
	}
	if spec.ClosedTag != "" {
		ct := spec.ClosedTag
		rt.ClosedTag = &ct
	}

	return rt
}
