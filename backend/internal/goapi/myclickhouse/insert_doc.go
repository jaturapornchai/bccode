package myclickhouse

import (
	"context"
	"fmt"

	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/myglobal"
)

func InsertDocListToClickHouse(ctx context.Context, holdingCode string, data []models.DocStruct, docRefData []models.DocRefStruct, docPaymentData []models.DocPaymentStruct) error {
	if len(data) == 0 && len(docRefData) == 0 && len(docPaymentData) == 0 {
		return nil
	}

	// DELETE-before-INSERT pattern
	if len(data) > 0 {
		DocDeleteClickHouse(ctx, holdingCode, data[0].DocNo)
	}

	connClickHouse, err := ClickHouseFastConnect()
	if err != nil {
		return fmt.Errorf("clickhouse connect: %w", err)
	}

	// 1. Insert Doc Records
	if len(data) > 0 {
		insertBatch, err := connClickHouse.PrepareBatch(ctx, fmt.Sprintf(`
			INSERT INTO %s (
				holdingcode, branchid, docno, docdatetime, perioddatetime,
				totalamount, paycashamount, paycashchange, paycashbalance,
				roundamount, checksum, slipurl, salechannelcode, deliveryamount,
				guidfixed, iscancel, cancelreason, guidpos, guidbranch, transflag,
				currency, currencysymbol,
				doccurrency, doccurrencysymbol,
				exchangerate, totalamountdoc,
				isdelete, approvalstatus, custcode
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, TableName("doc")))
		if err == nil {
			defer insertBatch.Close()
			for _, item := range data {
				insertBatch.Append(
					holdingCode, item.BranchID, item.DocNo, item.DocDateTime, item.PeriodDateTime,
					item.TotalAmount, item.PayCashAmount, item.PayCashChange, item.PayCashBalance,
					item.RoundAmount, item.Checksum, item.SlipURL, item.SaleChannelCode, item.DeliveryAmount,
					item.GuidFixed, item.IsCancel, item.CancelReason, item.GuidPOS, item.GuidBranch, item.TransFlag,
					// Multi-Currency Fields
					item.Currency, item.CurrencySymbol,
					item.DocCurrency, item.DocCurrencySymbol,
					item.ExchangeRate, item.TotalAmountDoc,
					// Soft Delete
					item.IsDelete,
					// สถานะการอนุมัติ
					item.ApprovalStatus,
					item.CustCode,
				)
			}
			insertBatch.Send()
		}
	}

	// 2. Insert DocRef Records
	if len(docRefData) > 0 {
		insertBatch, err := connClickHouse.PrepareBatch(ctx, fmt.Sprintf(`
			INSERT INTO %s (
				holdingcode, docno, docnotransflag, docnoref, docnoreftransflag
			) VALUES (?, ?, ?, ?, ?)
		`, TableName("docref")))
		if err == nil {
			defer insertBatch.Close()
			for _, item := range docRefData {
				insertBatch.Append(holdingCode, item.DocNo, item.DocNoTransFlag, item.DocRefNo, item.DocRefNoTransFlag)
			}
			insertBatch.Send()
		}
	}

	// 3. Insert DocPayment Records
	if len(docPaymentData) > 0 {
		insertBatch, err := connClickHouse.PrepareBatch(ctx, fmt.Sprintf(`
			INSERT INTO %s (
				holdingcode, branchid, docdatetime, perioddatetime, amount,
				description, docno, transflag, guidfixed, guidbranch
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, TableName("docpayment")))
		if err == nil {
			defer insertBatch.Close()
			for _, item := range docPaymentData {
				amountDecimal := fmt.Sprintf("%.6f", item.Amount)
				insertBatch.Append(
					holdingCode, item.BranchID, item.DocDateTime, item.PeriodDateTime, amountDecimal,
					item.Description, item.DocNo, item.TransFlag, item.GuidFixed, item.GuidBranch,
				)
			}
			insertBatch.Send()
		}
	}

	return nil
}

func InsertDocDetailListToClickHouse(ctx context.Context, holdingCode string, data []models.DocDetailStruct) error {
	if len(data) == 0 {
		return nil
	}

	connClickHouse, err := ClickHouseFastConnect()
	if err != nil {
		return fmt.Errorf("clickhouse connect: %w", err)
	}

	// Build barcode lookup map
	barcodeSet := make(map[string]struct{})
	for _, item := range data {
		if item.Barcode != "" {
			barcodeSet[item.Barcode] = struct{}{}
		}
	}

	barcodeList := make(map[string]models.ProductDocRefStruct)
	if len(barcodeSet) > 0 {
		barcodeListForQuery := ""
		for barcode := range barcodeSet {
			if barcodeListForQuery != "" {
				barcodeListForQuery += ","
			}
			barcodeListForQuery += fmt.Sprintf("'%s'", barcode)
		}

		query := fmt.Sprintf(`SELECT barcode,itemcode,unitcode,unitstand,unitdivide FROM %s WHERE holdingcode = ? and barcode in (`, TableName("productbarcode")) + barcodeListForQuery + `) order by barcode`
		rows, err := connClickHouse.Query(ctx, query, holdingCode)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var barcode, itemCode, unitCode string
				var unitStand, unitDivide float64
				if rows.Scan(&barcode, &itemCode, &unitCode, &unitStand, &unitDivide) == nil {
					barcodeList[barcode] = models.ProductDocRefStruct{
						Barcode: barcode, ItemCode: itemCode, UnitCode: unitCode,
						UnitStand: unitStand, UnitDivide: unitDivide,
					}
				}
			}
		}
	}

	// Prepare batch insert
	insertBatch, err := connClickHouse.PrepareBatch(ctx, fmt.Sprintf(`
		INSERT INTO %s (
			holdingcode, branchid, docno, docdatetime, perioddatetime,
			linenumber, barcode, barcodemain, qty, price, sumamount, discountamount,
			itemname, itemnames, refguid, sumamountchoice, ischoice, guidfixed, guidpos,
			guidbranch, transflag, itemcode, unitcode, unitstand, unitdivide, calcflag, calcseq, iscalcstock,
			pricedoc, sumamountdoc,
			discountamountdoc, priceexcludevatdoc, sumamountexcludevatdoc, totalvaluevatdoc
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, TableName("docdetail")))
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}
	defer insertBatch.Close()

	for _, docDetail := range data {
		if docDetail.DocNo == "" {
			continue
		}

		// Get values
		itemCode := docDetail.ItemCode
		unitStand := docDetail.UnitStand
		unitDivide := docDetail.UnitDivide
		barcode := docDetail.Barcode

		if ref, ok := barcodeList[barcode]; ok {
			if ref.ItemCode != "" {
				itemCode = ref.ItemCode
			}
			unitStand = ref.UnitStand
			unitDivide = ref.UnitDivide
		}

		if barcode == "" {
			barcode = itemCode
			if barcode == "" {
				barcode = "UNKNOWN"
			}
		}
		if unitStand == 0 {
			unitStand = 1.0
		}
		if unitDivide == 0 {
			unitDivide = 1.0
		}

		itemName := docDetail.Description
		if itemName == "" {
			itemName = itemCode
			if itemName == "" {
				itemName = "UNKNOWN"
			}
		}

		insertBatch.Append(
			holdingCode, "00000", docDetail.DocNo, docDetail.DocDateTime, docDetail.DocDateTime,
			docDetail.LineNumber, barcode, docDetail.BarcodeMain, docDetail.TotalQty, docDetail.Price,
			docDetail.SumAmount, 0.0, itemName, itemName, "", docDetail.SumAmount, 0, "", "", "",
			docDetail.TransFlag, itemCode, docDetail.UnitCode, unitStand, unitDivide,
			int(myglobal.GetTransactionMultiplier(docDetail.TransFlag)), docDetail.CalcSeq,
			myglobal.IsCalcStockTransFlag(docDetail.TransFlag),
			docDetail.PriceDoc, docDetail.SumAmountDoc,
			docDetail.DiscountAmountDoc, docDetail.PriceExcludeVatDoc,
			docDetail.SumAmountExcludeVatDoc, docDetail.TotalValueVatDoc,
		)
	}

	if err := insertBatch.Send(); err != nil {
		return fmt.Errorf("send batch: %w", err)
	}

	return nil
}
