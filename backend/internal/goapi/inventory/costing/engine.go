// Package costing คำนวณต้นทุนสินค้าด้วยวิธีถัวเฉลี่ยเคลื่อนที่ (moving weighted average) วิธีเดียว
// ตามคำสั่งเจ้าของ 2026-10-11 (ADR docs/kms/decisions/2026-10-11-stock-cost-average-only.md)
// ห้ามเพิ่มวิธีอื่น (FIFO, LIFO, FEFO, lot/specific, standard) กลับเข้ามา
package costing

import (
	"context"
	"database/sql"
	"fmt"

	inv "smlcloudplatform/internal/goapi/inventory/models"
)

// === Helper Functions ของ engine ===

// getOrCreateBalance — ดึง stock balance ตามบัญชี หรือสร้างใหม่ถ้ายังไม่มี
func getOrCreateBalance(ctx context.Context, tx *sql.Tx, holdingCode, itemCode, barcode, whCode, locationCode string) (*inv.InventoryStockBalance, error) {
	var balance inv.InventoryStockBalance
	err := tx.QueryRowContext(ctx,
		`SELECT id, holdingcode, itemcode, barcode, whcode, locationcode,
		        currentqty, currentavgcost, currenttotalvalue,
		        lastpurchasecost, lastpurchasedate, updatedat
		 FROM inventorystockbalances
		 WHERE holdingcode = $1 AND itemcode = $2 AND whcode = $3 AND locationcode = $4
		 FOR UPDATE`,
		holdingCode, itemCode, whCode, locationCode,
	).Scan(
		&balance.ID, &balance.HoldingCode, &balance.ItemCode, &balance.Barcode,
		&balance.WhCode, &balance.LocationCode,
		&balance.CurrentQty, &balance.CurrentAvgCost, &balance.CurrentTotalValue,
		&balance.LastPurchaseCost, &balance.LastPurchaseDate, &balance.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		// สร้าง balance ใหม่
		balance = inv.InventoryStockBalance{
			HoldingCode:  holdingCode,
			ItemCode:     itemCode,
			Barcode:      barcode,
			WhCode:       whCode,
			LocationCode: locationCode,
		}
		err = tx.QueryRowContext(ctx,
			`INSERT INTO inventorystockbalances
			 (holdingcode, itemcode, barcode, whcode, locationcode, currentqty, currentavgcost, currenttotalvalue, updatedat)
			 VALUES ($1, $2, $3, $4, $5, 0, 0, 0, NOW())
			 RETURNING id`,
			holdingCode, itemCode, barcode, whCode, locationCode,
		).Scan(&balance.ID)
		if err != nil {
			return nil, fmt.Errorf("สร้าง stock balance ไม่สำเร็จ: %w", err)
		}
		return &balance, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ดึง stock balance ไม่สำเร็จ: %w", err)
	}
	return &balance, nil
}

// updateBalance — อัพเดท stock balance
func updateBalance(ctx context.Context, tx *sql.Tx, balance *inv.InventoryStockBalance) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE inventorystockbalances
		 SET currentqty = $1, currentavgcost = $2, currenttotalvalue = $3,
		     lastpurchasecost = $4, lastpurchasedate = $5, updatedat = NOW()
		 WHERE id = $6`,
		balance.CurrentQty, balance.CurrentAvgCost, balance.CurrentTotalValue,
		balance.LastPurchaseCost, balance.LastPurchaseDate, balance.ID,
	)
	if err != nil {
		return fmt.Errorf("อัพเดท stock balance ไม่สำเร็จ: %w", err)
	}
	return nil
}

// insertCostTransaction — บันทึก cost transaction
func insertCostTransaction(ctx context.Context, tx *sql.Tx, ct *inv.InventoryCostTransaction) error {
	return tx.QueryRowContext(ctx,
		`INSERT INTO inventorycosttransactions
		 (holdingcode, itemcode, barcode, whcode, locationcode, transactiontype, transflag,
		  refdoctype, refdocno, qty, unitcost, totalcost, landedcost, lotnumber, expirydate,
		  balanceqty, balanceavgcost, balancetotalvalue,
		  transactiondate, accountingperiod, createdby, createdat, isreversed)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,NOW(),false)
		 RETURNING id`,
		ct.HoldingCode, ct.ItemCode, ct.Barcode, ct.WhCode, ct.LocationCode,
		ct.TransactionType, ct.TransFlag, ct.RefDocType, ct.RefDocNo,
		ct.Qty, ct.UnitCost, ct.TotalCost, ct.LandedCost,
		ct.LotNumber, ct.ExpiryDate,
		ct.BalanceQty, ct.BalanceAvgCost, ct.BalanceTotalValue,
		ct.TransactionDate, ct.AccountingPeriod, ct.CreatedBy,
	).Scan(&ct.ID)
}

// averageCostCalc — คำนวณต้นทุนเฉลี่ย (ป้องกัน divide by zero)
func averageCostCalc(totalValue, totalQty float64) float64 {
	if totalQty <= 0 {
		return 0
	}
	return totalValue / totalQty
}
