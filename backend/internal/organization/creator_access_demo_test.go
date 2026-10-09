package organization

import (
	"context"
	"testing"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"

	micromodels "smlcloudplatform/pkg/microservice/models"
)

// บัญชี Demo ไม่มี Google identity: สร้าง Holding ใหม่ไม่ได้ (ไม่งั้นทุกคนที่กดปุ่ม Demo สร้าง tenant ว่างถาวร)
// แต่ยังจัดการบริษัท/สาขาใน Holding ที่ seed ไว้ให้ได้
func TestDemoAccountCannotBootstrapHolding(t *testing.T) {
	t.Setenv("BCAI_DEMO_LOGIN_ENABLED", "true")
	t.Setenv("BCAI_DEMO_USERNAME", "demo")
	user := micromodels.UserInfo{UID: "uid-1", HoldingCode: "rungrueng"}

	for _, tc := range []struct {
		name         string
		username     string
		hasGoogle    bool
		requireAdmin bool
		wantGoogle   bool // want errGoogleRequired
	}{
		{"demo cannot create a holding", "demo", false, false, true},
		{"demo may still manage companies and branches", "demo", false, true, false},
		{"ordinary account without Google cannot create a holding", "somchai", false, false, true},
		{"ordinary account without Google cannot create companies", "somchai", false, true, true},
		{"demo with a real Google identity is unaffected", "demo", true, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			mock.ExpectQuery("FROM users u").WithArgs("uid-1").
				WillReturnRows(sqlmock.NewRows([]string{"username", "is_active", "exists"}).AddRow(tc.username, true, tc.hasGoogle))

			got := requireOrganizationCreator(context.Background(), db, user, tc.requireAdmin)
			if (got == errGoogleRequired) != tc.wantGoogle {
				t.Fatalf("requireOrganizationCreator(%s, admin=%v) = %v, want errGoogleRequired=%v", tc.username, tc.requireAdmin, got, tc.wantGoogle)
			}
			if !tc.requireAdmin && !tc.wantGoogle && got != nil {
				t.Fatalf("holding bootstrap should pass for a Google-linked account, got %v", got)
			}
		})
	}
}
