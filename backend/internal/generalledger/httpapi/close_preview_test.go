package httpapi

import (
	"testing"

	gl "smlcloudplatform/internal/generalledger"
)

// The closing-entries preview saves nothing, so the close screen's read permission is enough; creating the
// drafts (action "close") still needs the update permission, and an unknown process action stays denied.
func TestClosePreviewNeedsOnlyCloseScreenRead(t *testing.T) {
	preview := gl.Command{Resource: "processes", Action: gl.ClosePreviewAction}
	if !calculationOnly(preview) || processScreens[preview.Action] != "financial-close" {
		t.Fatalf("preview must be calculation-only on the close screen: %q", processScreens[preview.Action])
	}
	if calculationOnly(gl.Command{Resource: "processes", Action: "close"}) || calculationOnly(gl.Command{Resource: "processes", Action: "year-end"}) || calculationOnly(gl.Command{Resource: "journals", Action: gl.ClosePreviewAction}) {
		t.Fatal("only processes/close-preview may skip the update permission")
	}
	if !allowed(map[string]bool{"financial-close": true}, processScreens[preview.Action], "") {
		t.Fatal("close screen read permission must allow the preview")
	}
	if allowed(map[string]bool{"gl-year-end": true}, processScreens[preview.Action], "") || allowed(map[string]bool{}, processScreens[preview.Action], "") {
		t.Fatal("preview allowed without the close screen permission")
	}
	if processScreens["unknown"] != "" {
		t.Fatal("unknown process action must have no screen")
	}
}
