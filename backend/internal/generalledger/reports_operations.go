package generalledger

import (
	"context"
	"fmt"
)

func (r reportContext) dailyCheck(ctx context.Context) (Report, error) {
	cte := `WITH journals AS (SELECT id,code,payload FROM gl_records WHERE company=$1 AND kind='journals' AND NOT COALESCE((payload->>'isdeleted')::boolean,false) AND payload->>'fiscalyear'=$2 AND (payload->>'date')::date BETWEEN $3::date AND $4::date AND ($6='' OR payload->>'branchcode'=$6) AND ($9='' OR payload->>'bookcode'=$9)), sums AS (
      SELECT j.id,j.code,j.payload,SUM((l->>'debit')::numeric) AS debit,SUM((l->>'credit')::numeric) AS credit FROM journals j CROSS JOIN LATERAL jsonb_array_elements(j.payload->'lines') l WHERE ($5='' OR l->>'accountcode'=$5) AND ($7='' OR l->>'departmentcode'=$7) AND ($8='' OR l->>'projectcode'=$8) GROUP BY j.id,j.code,j.payload
    ), result AS (SELECT id,code AS docno,payload->>'date' AS date,payload->>'bookcode' AS bookcode,payload->>'status' AS status,payload->>'description' AS description,debit::text,credit::text,(debit-credit)::text AS difference,(CASE WHEN debit<>credit THEN 'ยอดไม่สมดุล' WHEN payload->>'status'='draft' THEN 'ยังไม่ผ่านรายการ' WHEN payload->>'status'='reversed' THEN 'กลับบัญชีแล้ว ต้นฉบับยังอยู่ในบัญชี' WHEN payload->>'status'='void' THEN 'ยกเลิกร่าง' ELSE 'ผ่านรายการแล้ว' END) AS checkresult FROM sums)`
	result, err := r.run(ctx, cte, "date,docno,id", []ReportColumn{textColumn("id", "รหัสรายการ"), textColumn("date", "วันที่"), textColumn("docno", "เลขที่เอกสาร"), textColumn("bookcode", "สมุดรายวัน"), textColumn("description", "รายละเอียด"), textColumn("status", "สถานะ"), amountColumn("debit", "เดบิต"), amountColumn("credit", "เครดิต"), amountColumn("difference", "ผลต่าง"), textColumn("checkresult", "ผลตรวจ")}, []string{"debit", "credit", "difference"})
	if r.query.AccountCode != "" || r.query.DepartmentCode != "" || r.query.ProjectCode != "" {
		result.Warnings = append(result.Warnings, "ยอดที่กรองตามบัญชี แผนก หรือโครงการเป็นเพียงบางบรรทัดของเอกสาร ผลต่างอาจไม่เป็นศูนย์")
	}
	return result, err
}

func (r reportContext) cashCTE() string {
	return r.base() + `, cash AS (SELECT * FROM filtered WHERE is_cash)`
}

func (r reportContext) cashFlow(ctx context.Context) (Report, error) {
	cte := r.cashCTE() + `, result AS (SELECT CASE cash_flow WHEN 'operating' THEN 'operating' WHEN 'investing' THEN 'investing' WHEN 'financing' THEN 'financing' ELSE 'unclassified' END AS category,
      CASE cash_flow WHEN 'operating' THEN 'กิจกรรมดำเนินงาน' WHEN 'investing' THEN 'กิจกรรมลงทุน' WHEN 'financing' THEN 'กิจกรรมจัดหาเงิน' ELSE 'ยังไม่ระบุประเภทกระแสเงินสด' END AS categoryname,SUM(debit)::text AS receipts,SUM(credit)::text AS payments,SUM(debit-credit)::text AS net
      FROM cash WHERE entry_date>=$3::date AND kind NOT IN ('opening','closing') GROUP BY cash_flow)`
	result, err := r.run(ctx, cte, "category", []ReportColumn{textColumn("category", "ประเภท"), textColumn("categoryname", "กิจกรรม"), amountColumn("receipts", "เงินรับ"), amountColumn("payments", "เงินจ่าย"), amountColumn("net", "เงินสดสุทธิ")}, []string{"receipts", "payments", "net"})
	if err != nil {
		return result, err
	}
	err = r.setTotals(ctx, &result, r.cashCTE()+` SELECT COALESCE(SUM(debit-credit) FILTER(WHERE entry_date<$3::date OR kind='opening'),0)::text,COALESCE(SUM(debit-credit),0)::text,(COUNT(*) FILTER(WHERE entry_date>=$3::date AND kind NOT IN ('opening','closing') AND cash_flow NOT IN ('operating','investing','financing')))::text FROM cash`, "opening", "closing", "unclassifiedlines")
	if err != nil {
		return result, err
	}
	if result.Totals["unclassifiedlines"] != "0" {
		result.Warnings = append(result.Warnings, "มีรายการเงินสดที่ยังไม่กำหนดกิจกรรม กรุณาจัดประเภทก่อนใช้งบกระแสเงินสด")
	}
	return result, nil
}

func (r reportContext) forecastCTE() string {
	return r.cashCTE() + `, opening AS (SELECT COALESCE(SUM(debit-credit),0) AS amount FROM cash WHERE entry_date<$3::date OR kind='opening'), plans AS (
      SELECT id,code,payload->>'name' AS name,(payload->>'startdate')::date AS date,payload->>'direction' AS direction,(payload->>'amount')::numeric AS amount FROM gl_records WHERE company=$1 AND kind='forecast' AND NOT COALESCE((payload->>'isdeleted')::boolean,false) AND COALESCE((payload->>'isactive')::boolean,false) AND payload->>'fiscalyear'=$2 AND (payload->>'startdate')::date BETWEEN $3::date AND $4::date AND ($5='' OR payload->>'accountcode'=$5) AND ($6='' OR payload->>'branchcode'=$6) AND ($7='' OR payload->>'departmentcode'=$7) AND ($8='' OR payload->>'projectcode'=$8) AND ($9='' OR payload->>'bookcode'=$9))`
}

func (r reportContext) cashFlowForecast(ctx context.Context) (Report, error) {
	cte := r.forecastCTE() + `, result AS (SELECT id,code,name,date::text,direction,(CASE WHEN direction='in' THEN amount ELSE 0 END)::text AS receipts,(CASE WHEN direction='out' THEN amount ELSE 0 END)::text AS payments,(CASE WHEN direction='in' THEN amount ELSE -amount END)::text AS net,((SELECT amount FROM opening)+SUM(CASE WHEN direction='in' THEN amount ELSE -amount END) OVER(ORDER BY date,code,id ROWS UNBOUNDED PRECEDING))::text AS balance FROM plans)`
	result, err := r.run(ctx, cte, "date,code,id", []ReportColumn{textColumn("date", "วันที่คาดการณ์"), textColumn("code", "รหัสแผน"), textColumn("name", "รายการ"), textColumn("direction", "ทิศทาง"), amountColumn("receipts", "คาดว่าจะรับ"), amountColumn("payments", "คาดว่าจะจ่าย"), amountColumn("net", "สุทธิ"), amountColumn("balance", "เงินสดคาดการณ์")}, []string{"receipts", "payments", "net"})
	if err != nil {
		return result, err
	}
	err = r.setTotals(ctx, &result, r.forecastCTE()+` SELECT (SELECT amount FROM opening)::text,((SELECT amount FROM opening)+COALESCE(SUM(CASE WHEN direction='in' THEN amount ELSE -amount END),0))::text FROM plans`, "opening", "closing")
	result.Warnings = append(result.Warnings, "ประมาณการจากแผนรับจ่ายที่ผู้ใช้บันทึก รวมยอดเงินสดจริงก่อนวันเริ่มต้น ไม่ใช่ยอดรับจ่ายที่เกิดขึ้นแล้ว")
	return result, err
}

func (r reportContext) dimensionProfit(ctx context.Context, name string) (Report, error) {
	dimension := `project_code`
	key := "projectcode"
	label := "โครงการ"
	if name == "dimensionpnl" {
		dimension = `branch_code`
		key = "branchcode"
		label = "สาขา"
	}
	group := dimension
	departmentSelect := ""
	columns := []ReportColumn{textColumn(key, label)}
	order := key
	if name == "dimensionpnl" {
		group += `,department_code`
		departmentSelect = `,department_code AS departmentcode`
		columns = append(columns, textColumn("departmentcode", "แผนก"))
		order += `,departmentcode`
	}
	cte := r.base() + `, grouped AS (SELECT ` + dimension + ` AS ` + key + departmentSelect + `,COALESCE(SUM(credit-debit) FILTER(WHERE account_type='income'),0) AS revenue,COALESCE(SUM(debit-credit) FILTER(WHERE account_type='expense'),0) AS expense,COUNT(DISTINCT journal_id) AS journals FROM filtered WHERE entry_date>=$3::date AND kind NOT IN ('opening','closing') AND account_type IN ('income','expense') GROUP BY ` + group + `), result AS (SELECT ` + key
	if name == "dimensionpnl" {
		cte += `,departmentcode`
	}
	cte += `,revenue::text,expense::text,(revenue-expense)::text AS profit,journals::text FROM grouped)`
	columns = append(columns, amountColumn("revenue", "รายได้"), amountColumn("expense", "ค่าใช้จ่าย"), amountColumn("profit", "กำไรขาดทุน"), textColumn("journals", "จำนวนเอกสาร"))
	result, err := r.run(ctx, cte, order, columns, []string{"revenue", "expense", "profit"})
	if err != nil {
		return result, err
	}
	if name == "projectsummary" {
		result.Warnings = append(result.Warnings, "สรุปเฉพาะรายได้และค่าใช้จ่ายที่ผ่านบัญชีและระบุโครงการ ไม่รวมประมาณการต้นทุนหรือความคืบหน้างาน")
	}
	for _, row := range result.Rows {
		if row[key] == "" {
			row[key] = "ยังไม่ระบุ"
		}
		if name == "dimensionpnl" && row["departmentcode"] == "" {
			row["departmentcode"] = "ยังไม่ระบุ"
		}
	}
	return result, nil
}

func (r reportContext) summary(ctx context.Context, name string) (Report, error) {
	cte := r.base() + `, monthly AS (SELECT to_char(entry_date,'YYYY-MM') AS month,COALESCE(SUM(credit-debit) FILTER(WHERE account_type='income'),0) AS revenue,COALESCE(SUM(debit-credit) FILTER(WHERE account_type='expense'),0) AS expense,COALESCE(SUM(debit-credit) FILTER(WHERE is_cash),0) AS cashchange FROM filtered WHERE entry_date>=$3::date AND kind NOT IN ('opening','closing') GROUP BY to_char(entry_date,'YYYY-MM')), result AS (SELECT month,revenue::text,expense::text,(revenue-expense)::text AS profit,cashchange::text FROM monthly)`
	result, err := r.run(ctx, cte, "month", []ReportColumn{textColumn("month", "เดือน"), amountColumn("revenue", "รายได้"), amountColumn("expense", "ค่าใช้จ่าย"), amountColumn("profit", "กำไรขาดทุน"), amountColumn("cashchange", "เงินสดเปลี่ยนแปลง")}, []string{"revenue", "expense", "profit", "cashchange"})
	if err != nil {
		return result, err
	}
	err = r.setTotals(ctx, &result, r.base()+` SELECT COALESCE(SUM(debit-credit) FILTER(WHERE is_cash),0)::text,COALESCE(SUM(debit-credit) FILTER(WHERE account_type='asset'),0)::text,COALESCE(SUM(credit-debit) FILTER(WHERE account_type='liability'),0)::text,COALESCE(SUM(credit-debit) FILTER(WHERE account_type IN ('equity','income','expense')),0)::text,COALESCE(SUM(debit-credit),0)::text FROM filtered`, "cash", "assets", "liabilities", "equity", "difference")
	if err == nil && name == "executivesummary" {
		result.Warnings = append(result.Warnings, fmt.Sprintf("ข้อมูลจากรายการผ่านบัญชี ปี %s สกุลเงิน %s ไม่รวมเอกสารร่างหรือการวิเคราะห์จากข้อมูลภายนอก", r.fiscal.Code, r.fiscal.Currency))
	}
	return result, err
}
