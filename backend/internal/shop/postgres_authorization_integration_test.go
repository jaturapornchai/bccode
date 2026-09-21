//go:build integration

package shop

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"smlcloudplatform/internal/authentication/models"
	msmodels "smlcloudplatform/pkg/microservice/models"
)

func TestPostgresMembershipFailsClosed(t *testing.T) {
	dsn := os.Getenv("GL_AUTH_TEST_DSN")
	if dsn == "" {
		t.Skip("GL_AUTH_TEST_DSN required")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	ns := fmt.Sprintf("authaudit_shop_%d", time.Now().UnixNano())
	if _, err = db.Exec("CREATE SCHEMA " + ns + "; SET search_path TO " + ns); err != nil {
		t.Fatal(err)
	}
	defer db.Exec("DROP SCHEMA " + ns + " CASCADE")
	_, err = db.Exec(`CREATE TABLE users(id uuid PRIMARY KEY,username text,is_active boolean);CREATE TABLE holdings(code text,name text,is_active boolean);
 CREATE TABLE holding_members(holding_code text,user_id uuid,role text,permission_sets jsonb,access_scopes jsonb,is_active boolean);
 CREATE TABLE companies(holding_code text,code text,is_active boolean);INSERT INTO holdings VALUES('H','Holding',true),('OTHER','Other',true);INSERT INTO companies VALUES('H','C',true);`)
	if err != nil {
		t.Fatal(err)
	}
	uid := uuid.NewString()
	if _, err = db.Exec(`INSERT INTO users VALUES($1,'member',true)`, uid); err != nil {
		t.Fatal(err)
	}
	repo := &ShopUserPostgresRepository{db: db}
	svc := NewShopUserService(repo)
	checkDenied := func() {
		t.Helper()
		if _, e := repo.FindByHoldingCodeAndUserUID(context.Background(), "H", uid); e == nil {
			t.Fatal("missing/inactive membership allowed")
		}
		if e := svc.SaveUserFullProfile("H", "member", &models.UserRoleRequest{Username: "member", Role: models.ROLE_OWNER}); e == nil {
			t.Fatal("self promotion allowed")
		}
		list, _, e := repo.FindByUserUIDPage(context.Background(), uid, msmodels.Pageable{})
		if e != nil || len(list) != 0 {
			t.Fatal("inaccessible Holding listed")
		}
	}
	checkDenied()
	if _, err = db.Exec(`INSERT INTO holding_members VALUES('H',$1,'owner','[]','[]',false)`, uid); err != nil {
		t.Fatal(err)
	}
	checkDenied()
	if _, err = db.Exec(`UPDATE holding_members SET is_active=true`); err != nil {
		t.Fatal(err)
	}
	u, err := repo.FindByHoldingCodeAndUserUID(context.Background(), "H", uid)
	if err != nil || u.Role != models.ROLE_OWNER || len(u.AccessScopes) != 1 {
		t.Fatal("real active owner lost scope")
	}
	if _, err = repo.FindByHoldingCodeAndUserUID(context.Background(), "H", "member"); err == nil {
		t.Fatal("username accepted as trusted UID")
	}
	list, _, err := repo.FindByUserUIDPage(context.Background(), uid, msmodels.Pageable{})
	if err != nil || len(list) != 1 || list[0].HoldingCode != "H" || list[0].Role != models.ROLE_OWNER {
		t.Fatal("Holding list not membership scoped")
	}
	if _, err = db.Exec(`UPDATE users SET is_active=false`); err != nil {
		t.Fatal(err)
	}
	checkDenied()
	if _, err = db.Exec(`UPDATE users SET is_active=true;UPDATE holdings SET is_active=false WHERE code='H'`); err != nil {
		t.Fatal(err)
	}
	checkDenied()
}
