package generalledger

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

func (s *Store) accountMasterReferences(ctx context.Context, scope Scope, code string) error {
	fields := []string{"accountcode", "rules.accountcode", "itemaccount", "costaccount", "revenueaccount", "profitlossaccount", "retainedearningsaccount", "rows.accountcodes"}
	conditions := make(bson.A, 0, len(fields))
	for _, field := range fields {
		conditions = append(conditions, bson.M{field: code})
	}
	collections := []string{"fiscal_year"}
	for _, name := range MasterCollections {
		collections = append(collections, name)
	}
	for _, name := range collections {
		f := scopeFilter(scope)
		f["isdeleted"] = false
		f["$or"] = conditions
		n, err := s.db.Collection(name).CountDocuments(ctx, f)
		if err != nil {
			return err
		}
		if n > 0 {
			return userError(CodeReferencedMaster, "บัญชีนี้ถูกใช้ในการตั้งค่าหรือแผนบัญชี กรุณาแก้รายการอ้างอิงก่อนลบ")
		}
	}
	return nil
}

func (s *Store) masterDeleteReferences(ctx context.Context, scope Scope, master Master) error {
	if master.Kind == "periods" {
		used, err := s.referenced(ctx, scope, bson.M{"fiscalyear": master.FiscalYear, "date": bson.M{"$gte": master.StartDate, "$lte": master.EndDate}, "isdeleted": false})
		if err != nil {
			return err
		}
		if used {
			return fmt.Errorf("งวดนี้มีรายการบัญชีอ้างอิง ลบไม่ได้")
		}
	}
	if master.Kind == "account-groups" {
		f := scopeFilter(scope)
		f["isdeleted"] = false
		f["accountgroup"] = master.Code
		n, err := s.db.Collection("chart_of_accounts").CountDocuments(ctx, f)
		if err != nil {
			return err
		}
		if n > 0 {
			return fmt.Errorf("กลุ่มนี้มีผังบัญชีอ้างอิง ลบไม่ได้")
		}
	}
	return nil
}

func (s *Store) yearMasterReferences(ctx context.Context, scope Scope, year FiscalYear, next *FiscalYear) error {
	if next != nil && next.Scale != year.Scale {
		for _, kind := range []string{"budgets", "forecast"} {
			f := scopeFilter(scope)
			f["isdeleted"] = false
			f["fiscalyear"] = year.Code
			n, err := s.db.Collection(MasterCollections[kind]).CountDocuments(ctx, f)
			if err != nil {
				return err
			}
			if n > 0 {
				return fmt.Errorf("ปีบัญชีมีงบประมาณหรือประมาณการแล้ว เปลี่ยนสกุลเงินหรือทศนิยมไม่ได้")
			}
		}
	}
	for _, kind := range []string{"periods", "budgets", "forecast"} {
		f := scopeFilter(scope)
		f["isdeleted"] = false
		f["fiscalyear"] = year.Code
		if next != nil {
			f["$or"] = bson.A{bson.M{"startdate": bson.M{"$lt": next.StartDate}}, bson.M{"enddate": bson.M{"$gt": next.EndDate}}}
		}
		n, err := s.db.Collection(MasterCollections[kind]).CountDocuments(ctx, f)
		if err != nil {
			return err
		}
		if n > 0 {
			return fmt.Errorf("ปีบัญชีมีงวดหรือแผนอ้างอิงอยู่ กรุณาตรวจสอบช่วงวันและข้อมูลก่อน")
		}
	}
	return nil
}
