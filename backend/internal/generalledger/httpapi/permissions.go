package httpapi

// Process and filing screens need the same source reports for their previews.
// These grants are read-only; posting and configuration still require actions.
func canReadReport(p map[string]bool, name string) bool {
	if allowed(p, reportScreens[name], "") {
		return true
	}
	var screens []string
	switch name {
	case "workingpaper":
		screens = []string{"financial-close", "gl-recalculate-posted", "gl-reprocess", "xbrl-export"}
	case "trialbalance":
		screens = []string{"gl-year-end"}
	}
	for _, screen := range screens {
		if allowed(p, screen, "") {
			return true
		}
	}
	return false
}
