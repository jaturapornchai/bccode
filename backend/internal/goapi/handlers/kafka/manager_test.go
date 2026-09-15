package kafka

import (
	"testing"

	"smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/process/stockengine"
)

// เอกสารทุกประเภทที่กระทบสต็อกต้องมีผู้รับฟัง ไม่เช่นนั้นข้อมูลจะไม่ไหลเข้าตารางที่ใช้คิดต้นทุน
//
// เคยเกิดขึ้นจริง: ตัวจัดการถูกเขียนครบทั้ง 50 ตัว แต่สมัครรับจริงเพียง 14 ตัว
// โอนคลัง ปรับสต็อก รับ-เบิก-คืนสินค้า และยอดยกมา จึงไม่เคยถูกนำมาคิดต้นทุนเลย
func TestEveryStockMovingDocumentHasAConsumer(t *testing.T) {
	expected := map[int]string{
		myglobal.TransFlagsToProcess[0]:      "Stock Balance",
		TRANS_FLAG_PURCHASE:                  "Purchase",
		TRANS_FLAG_PURCHASE_PARTIAL:          "Purchase Partial",
		TRANS_FLAG_SALE_INVOICE_RETURN:       "Sale Return",
		TRANS_FLAG_STOCK_RECEIVE_PRODUCT:     "Stock Receive Product",
		TRANS_FLAG_STOCK_RETURN_PRODUCT:      "Stock Return Product",
		TRANS_FLAG_STOCK_ADJUSTMENT_INCREASE: "Stock Adjustment",
		TRANS_FLAG_SALE_INVOICE:              "Sale Invoice",
		TRANS_FLAG_PURCHASE_RETURN:           "Purchase Return",
		TRANS_FLAG_STOCK_PICKUP_PRODUCT:      "Stock Pickup Product",
		TRANS_FLAG_STOCK_ADJUSTMENT_DECREASE: "Stock Adjustment",
		TRANS_FLAG_STOCK_TRANSFER:            "Stock Transfer",
	}

	subscribed := map[string]bool{}
	for _, group := range consumerGroups() {
		subscribed[group.name] = true
	}

	for _, flag := range myglobal.TransFlagsToProcess {
		name, known := expected[flag]
		if !known {
			t.Errorf("transflag %d moves stock but this test does not say which consumer covers it", flag)
			continue
		}
		if !subscribed[name] {
			t.Errorf("transflag %d (%s) moves stock but no consumer group is subscribed for it", flag, name)
		}
	}
}

// ทุกรายการในตารางต้องกรอกครบ กันการเพิ่มรายการใหม่แบบตกหล่น
func TestConsumerGroupsAreFullyConfigured(t *testing.T) {
	seenGroups := map[string][]string{}

	for _, group := range consumerGroups() {
		if group.name == "" {
			t.Error("a consumer group has no name")
			continue
		}
		if group.group == "" {
			t.Errorf("%s has no consumer group id", group.name)
		}
		for label, topic := range map[string]string{
			"create": group.createTopic, "update": group.updateTopic, "delete": group.deleteTopic,
		} {
			if topic == "" {
				t.Errorf("%s has no %s topic", group.name, label)
			}
		}
		if group.onCreateOrUpdate == nil {
			t.Errorf("%s has no create/update handler", group.name)
		}
		if group.onDelete == nil {
			t.Errorf("%s has no delete handler", group.name)
		}
		seenGroups[group.group] = append(seenGroups[group.group], group.name)
	}

	// หัวข้อของเอกสารคนละประเภทต้องไม่ใช้กลุ่มเดียวกัน ไม่งั้นเอกสารที่ติดขัดจะฉุดเอกสารอื่นไปด้วย
	for id, names := range seenGroups {
		if len(names) > 1 {
			t.Errorf("consumer group %q is shared by %v; a stuck document would stall the others", id, names)
		}
	}
}

// ทิศทางของทุกประเภทเอกสารที่ระบบนำมาคิดสต็อก ต้องเป็นที่รู้จักของเครื่องคิดต้นทุน
func TestStockEngineKnowsEveryProcessedTransFlag(t *testing.T) {
	for _, flag := range myglobal.TransFlagsToProcess {
		if _, known := stockengine.DirectionOf(flag); !known {
			t.Errorf("transflag %d is processed for stock but the cost engine does not know its direction", flag)
		}
	}
}

// รายชื่อประเภทเอกสารสองชุดต้องเท่ากันเสมอ
//
// การคิดใหม่ลบสมุดสต็อกของสินค้าตั้งแต่ต้นงวดทุกประเภทเอกสาร แล้วเขียนกลับเฉพาะประเภทที่อยู่ในรายชื่อที่ส่งมา
// ถ้าเครื่องคิดต้นทุนรู้จักประเภทที่ไม่อยู่ในรายชื่อ ประวัติของประเภทนั้นจะหายไปเงียบ ๆ ทุกครั้งที่คิดใหม่
func TestProcessedTransFlagsCoverEveryDirectionTheEngineKnows(t *testing.T) {
	processed := map[int]bool{}
	for _, flag := range myglobal.TransFlagsToProcess {
		processed[flag] = true
	}

	for _, flag := range stockengine.KnownTransFlags() {
		if !processed[flag] {
			t.Errorf("the cost engine knows transflag %d but it is not in TransFlagsToProcess; its ledger rows would be deleted and never rewritten", flag)
		}
	}
}
