package generalledger

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// The ledger "คำอธิบาย" prefixes follow Champ's own order Dep, Proj, Allo, Part, Side, Job, Bran (GLRepGL2View.cpp:889-926),
// not the order of the registers on the screens.
func TestLedgerDescriptionFollowsChampPrefixOrder(t *testing.T) {
	last := -1
	for _, prefix := range []string{"Dep: ", "Proj: ", "Allo: ", "Part: ", "Side: ", "Job: ", "Bran: "} {
		at := strings.Index(ledgerDescription, "'"+prefix+"'")
		if at < 0 {
			t.Fatalf("prefix %q missing from the ledger description", prefix)
		}
		if at <= last {
			t.Fatalf("prefix %q is out of Champ order", prefix)
		}
		last = at
	}
}

// docNoFilter takes the two LAST arguments: the dimension arguments are appended before them, so a fixed $12/$13 would
// silently filter on a department list.
func TestDocNoFilterUsesTheLastTwoArguments(t *testing.T) {
	for _, n := range []int{13, 15, 19} {
		r := reportContext{args: make([]any, n)}
		filter := r.docNoFilter()
		from, to := "$"+strconv.Itoa(n-1), "$"+strconv.Itoa(n)
		if !strings.Contains(filter, from+"=''") || !strings.Contains(filter, "doc_no>="+from) || !strings.Contains(filter, "doc_no<="+to) {
			t.Fatalf("%d args: filter %q does not use %s and %s", n, filter, from, to)
		}
	}
}

// R9: "เลือกทั้งหมด" is ticked only where the flag is on and nothing says which codes are allowed.
func TestTickAllCodesOnlyForFlaggedAccountsWithoutAList(t *testing.T) {
	a := Account{
		DepartSeparate:  true,
		ProjectSeparate: true, ProjectCodes: []string{"P01"},
		JobSeparate: true, JobAllCodes: true,
		SideSeparate: false,
	}
	if got := tickAllCodes(&a); !reflect.DeepEqual(got, []string{DepartmentKind}) {
		t.Fatalf("ticked %v, want only the department", got)
	}
	if !a.DepartAllCodes || a.ProjectAllCodes || a.SideAllCodes || a.AllocateAllCodes || a.PartAllCodes {
		t.Fatalf("unexpected boxes after ticking: %+v", a)
	}
	if got := tickAllCodes(&a); len(got) != 0 {
		t.Fatalf("second pass ticked %v, want nothing (idempotent)", got)
	}
}

func TestDimensionProvisionPlanEmpty(t *testing.T) {
	if !(DimensionProvisionPlan{}).Empty() {
		t.Fatal("a plan without switches and accounts must be empty")
	}
	if (DimensionProvisionPlan{Switches: []string{JobKind}}).Empty() || (DimensionProvisionPlan{Accounts: []string{"1000"}}).Empty() {
		t.Fatal("a plan with work to do reported empty")
	}
}

// Year-end groups balances by branch + the six codes; the codes go back onto the generated line in the same slots.
func TestProcessBalanceDimensionCodesRoundTrip(t *testing.T) {
	balance := ProcessBalance{DepartmentCode: "D1", ProjectCode: "P1", AllocateCode: "AL1", JobCode: "J1", PartCode: "PT1", SideCode: "SD1"}
	codes := balance.dimensionCodes()
	if codes != (dimensionCodes{"D1", "P1", "AL1", "J1", "PT1", "SD1"}) {
		t.Fatalf("codes = %v, want registry order", codes)
	}
	var line Line
	codes.applyTo(&line)
	if line.DepartmentCode != "D1" || line.ProjectCode != "P1" || line.AllocateCode != "AL1" || line.JobCode != "J1" || line.PartCode != "PT1" || line.SideCode != "SD1" {
		t.Fatalf("line = %+v", line)
	}
	if got := len(balance.scanTargets()); got != 4+len(dimensionRegistry)+1 {
		t.Fatalf("%d scan targets, want account(3)+branch+%d codes+balance", got, len(dimensionRegistry))
	}
	if !(dimensionCodes{"D1", "P1"}).less(dimensionCodes{"D1", "P2"}) || (dimensionCodes{"D2"}).less(dimensionCodes{"D1", "P2"}) || (dimensionCodes{"D1"}).less(dimensionCodes{"D1"}) {
		t.Fatal("less must order by the first differing register")
	}
}
