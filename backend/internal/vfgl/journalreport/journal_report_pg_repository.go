package journalreport

import (
	"fmt"
	"smlcloudplatform/internal/vfgl/journalreport/models"
	"smlcloudplatform/pkg/microservice"
	"strconv"
	"time"

	"github.com/lib/pq"
)

type IJournalReportPgRepository interface {
	GetDataTrialBalance(holdingCode string, accountGroup string, includeCloseAccountMode bool, startDate time.Time, endDate time.Time) ([]models.TrialBalanceSheetAccountDetail, error)
	GetDataProfitAndLoss(holdingCode string, accountGroup string, includeCloseAccountMode bool, startDate time.Time, endDate time.Time) ([]models.ProfitAndLossSheetAccountDetail, error)
	GetDataBalanceSheet(holdingCode string, accountGroup string, includeCloseAccountMode bool, endDate time.Time) ([]models.BalanceSheetAccountDetail, error)
	GetDataLedgerAccount(holdingCode string, accountGroup string, creditorCode string, debtorCode string, consolidateAccountCode string, accountRanges []models.LedgerAccountCodeRange, bookCode string, startDate time.Time, endDate time.Time) ([]models.LedgerAccountRaw, error)
	GetMultiShopRevenue(holdingCodes []string, startDate time.Time, endDate time.Time) ([]models.MultiShopRevenueRaw, error)
}

type JournalReportPgRepository struct {
	pst microservice.IPersister
}

func NewJournalReportPgRepository(pst microservice.IPersister) JournalReportPgRepository {
	return JournalReportPgRepository{
		pst: pst,
	}
}

/* Full Query

-- REPORT TRIAL BALANCE SHEET

WITH journal_doc as (
		select h.holdingcode, h.docno, h.docdate, h.accountyear, h.accountperiod, h.accountgroup
			, d.accountcode
			--, acc.accountname , acc.accountcategory, acc.accountbalancetype
			, d.debitamount ,d.creditamount
			from journals_detail as d
			join journals as h on h.holdingcode = d.holdingcode and h.docno = d.docno
			where h.holdingcode= '27dcEdktOoaSBYFmnN6G6ett4Jb' and h.accountgroup = '01'
			--left join chartofaccounts as acc on acc.holdingcode = d.holdingcode and acc.accountcode = d.accountcode
		)
		, bal as (
			select accountcode, sum(debitamount) as debitamount, sum(creditamount) as creditamount
			from journal_doc where journal_doc.docdate < '2022-05-01'
			group by accountcode
		)
		, prd as (
			select accountcode, sum(debitamount) as debitamount, sum(creditamount) as creditamount
			from journal_doc where journal_doc.docdate between '2022-05-01' and '2022-05-31'
			group by accountcode
		)
		, nex as (
			select accountcode, sum(debitamount) as debitamount, sum(creditamount) as creditamount
			from journal_doc where journal_doc.docdate <= '2022-05-31'
			group by accountcode
		)
		, journal_sheet_sum as (
			select chart.holdingcode, chart.par_id
            , chart.accountcode, chart.accountname
			, chart.accountcategory, chart.accountbalancetype, chart.accountgroup, chart.accountlevel, chart.consolidateaccountcode
			, coalesce(bal.debitamount, 0) as balancedebitamount, coalesce(bal.creditamount, 0) as balancecreditamount
			, coalesce(prd.debitamount, 0) as debitamount, coalesce(prd.creditamount, 0) as creditamount
			, coalesce(nex.debitamount, 0) as nextbalancedebitamount, coalesce(nex.creditamount, 0) as nextbalancecreditamount
			, case when(accountbalancetype = 1) then coalesce(bal.debitamount, 0)-coalesce(bal.creditamount, 0)
				else coalesce(bal.creditamount, 0)-coalesce(bal.debitamount, 0)
				end as balanceamount
			, case when(accountbalancetype = 1) then coalesce(prd.debitamount, 0)-coalesce(prd.creditamount, 0)
				else coalesce(prd.creditamount, 0)-coalesce(prd.debitamount, 0)
				end as amount
			, case when(accountbalancetype = 1) then coalesce(nex.debitamount, 0)-coalesce(nex.creditamount, 0)
				else coalesce(nex.creditamount, 0)-coalesce(nex.debitamount, 0)
				end as nextbalanceamount
			from chartofaccounts as chart
			left join bal on bal.accountcode = chart.accountcode
			left join prd on prd.accountcode = chart.accountcode
			left join nex on nex.accountcode = chart.accountcode
			where chart.holdingcode= '27dcEdktOoaSBYFmnN6G6ett4Jb'
		)
		select holdingcode, par_id
        , accountcode, accountname, accountcategory, accountbalancetype
        , accountgroup, accountlevel, consolidateaccountcode
		, balancedebitamount, balancecreditamount, debitamount, creditamount, nextbalancedebitamount, nextbalancecreditamount
		, balanceamount, amount, nextbalanceamount
		from journal_sheet_sum
		where balanceamount <> 0 or amount <> 0 or nextbalanceamount <> 0
		order by accountcode

*/

func (repo JournalReportPgRepository) GetDataTrialBalance(holdingCode string, accountGroup string, includeCloseAccountMode bool,
	startDate time.Time, endDate time.Time) ([]models.TrialBalanceSheetAccountDetail, error) {

	var closeDocFilter string

	accountGroupFilter := ``
	if len(accountGroup) > 0 {
		accountGroupFilter = ` and h.accountgroup = @accountgroup `
	}

	if includeCloseAccountMode == true {
		closeDocFilter = `
		union all
		select h.holdingcode, h.docno, h.docdate, h.accountyear, h.accountperiod, h.accountgroup
			, d.accountcode
			, d.debitamount ,d.creditamount
			, acc.accountcategory
			from journals_detail as d
			join journals as h on h.holdingcode = d.holdingcode and h.docno = d.docno
			left join chartofaccounts as acc on acc.holdingcode = d.holdingcode and acc.accountcode = d.accountcode
            where h.holdingcode= @holdingcode and ( h.journaltype = 1 and h.docdate between @startdate and @enddate )
	`
	}
	query := `
	WITH journal_doc as (
		select h.holdingcode, h.docno, h.docdate, h.accountyear, h.accountperiod, h.accountgroup
			, d.accountcode
			--, acc.accountname , acc.accountbalancetype
			, d.debitamount ,d.creditamount
			, acc.accountcategory
			from journals_detail as d
			join journals as h on h.holdingcode = d.holdingcode and h.docno = d.docno
			left join chartofaccounts as acc on acc.holdingcode = d.holdingcode and acc.accountcode = d.accountcode
			where h.holdingcode= @holdingcode ` + accountGroupFilter + `  and h.docdate <= @enddate
			 and (( h.journaltype = 0) or (h.journaltype=1 and h.docdate < @startdate ))

		` + closeDocFilter + `

		)
		, bal as (
			select accountcode, sum(debitamount) as debitamount, sum(creditamount) as creditamount
			from journal_doc where journal_doc.docdate < @startdate

			group by accountcode
		)
		, prd as (
			select accountcode, sum(debitamount) as debitamount, sum(creditamount) as creditamount
			from journal_doc where journal_doc.docdate between @startdate and @enddate
			group by accountcode
		)
		, nex as (
			select accountcode, sum(debitamount) as debitamount, sum(creditamount) as creditamount
			from journal_doc
			where
			 ( journal_doc.docdate <= @enddate )
			group by accountcode
		)
		, journal_sheet_sum as (
			select chart.holdingcode, chart.parid
            , chart.accountcode, chart.accountname
			, chart.accountcategory, chart.accountbalancetype, chart.accountgroup, chart.accountlevel, chart.consolidateaccountcode
			, coalesce(bal.debitamount, 0) as balancedebitamount, coalesce(bal.creditamount, 0) as balancecreditamount
			, coalesce(prd.debitamount, 0) as debitamount, coalesce(prd.creditamount, 0) as creditamount
			, case when(accountcategory = 1 or accountcategory = 5) then coalesce(nex.debitamount, 0)-coalesce(nex.creditamount, 0) else 0 end as nextbalancedebitamount
            , case when(accountcategory = 1 or accountcategory = 5) then 0 else coalesce(nex.creditamount, 0)-coalesce(nex.debitamount, 0) end as nextbalancecreditamount
			, case when(accountcategory = 1 or accountcategory = 5) then coalesce(bal.debitamount, 0)-coalesce(bal.creditamount, 0)
				else coalesce(bal.creditamount, 0)-coalesce(bal.debitamount, 0)
				end as balanceamount
			, case when(accountcategory = 1 or accountcategory = 5) then coalesce(prd.debitamount, 0)-coalesce(prd.creditamount, 0)
				else coalesce(prd.creditamount, 0)-coalesce(prd.debitamount, 0)
				end as amount
			, case when(accountcategory = 1 or accountcategory = 5) then coalesce(nex.debitamount, 0)-coalesce(nex.creditamount, 0)
				else coalesce(nex.creditamount, 0)-coalesce(nex.debitamount, 0)
				end as nextbalanceamount
			from chartofaccounts as chart
			left join bal on bal.accountcode = chart.accountcode
			left join prd on prd.accountcode = chart.accountcode
			left join nex on nex.accountcode = chart.accountcode
			where chart.holdingcode= @holdingcode
		)
		select holdingcode, parid
        , accountcode, accountname, accountcategory, accountbalancetype
        , accountgroup, accountlevel, consolidateaccountcode
		, balancedebitamount, balancecreditamount, debitamount, creditamount, nextbalancedebitamount, nextbalancecreditamount
		, balanceamount, amount, nextbalanceamount
		from journal_sheet_sum
		where balanceamount <> 0 or amount <> 0 or nextbalanceamount <> 0
		order by accountcode
	`

	var details []models.TrialBalanceSheetAccountDetail

	condition := map[string]interface{}{
		"holdingcode": holdingCode,
		"start_date":  startDate,
		"end_date":    endDate,
	}

	if len(accountGroup) > 0 {
		condition["accountgroup"] = accountGroup
	}

	_, err := repo.pst.Raw(query, condition, &details)
	if err != nil {
		return nil, err
	}

	return details, nil
}

func (repo JournalReportPgRepository) GetDataProfitAndLoss(holdingCode string, accountGroup string, includeCloseAccountMode bool, startDate time.Time, endDate time.Time) ([]models.ProfitAndLossSheetAccountDetail, error) {

	var closeDocFilter string

	accountGroupFilter := ``
	if len(accountGroup) > 0 {
		accountGroupFilter = ` and h.accountgroup = @accountgroup `
	}

	if includeCloseAccountMode == true {
		closeDocFilter = `
		union all
		select h.holdingcode, h.docno, h.docdate, h.accountyear, h.accountperiod, h.accountgroup
			, d.accountcode
			, d.debitamount ,d.creditamount
			, acc.accountcategory
			from journals_detail as d
			join journals as h on h.holdingcode = d.holdingcode and h.docno = d.docno
			left join chartofaccounts as acc on acc.holdingcode = d.holdingcode and acc.accountcode = d.accountcode

            where h.holdingcode= @holdingcode and ( h.journaltype = 1 and h.docdate between @startdate and @enddate )
	`
	}

	query := `
	WITH journal_doc as (
		select h.holdingcode, h.docno, h.docdate, h.accountyear, h.accountperiod, h.accountgroup
			, d.accountcode
			, d.debitamount ,d.creditamount
			, acc.accountcategory
			from journals_detail as d
			join journals as h on h.holdingcode = d.holdingcode and h.docno = d.docno
			left join chartofaccounts as acc on acc.holdingcode = d.holdingcode and acc.accountcode = d.accountcode

			where h.holdingcode= @holdingcode ` + accountGroupFilter + ` and h.docdate < @enddate
			and (( h.journaltype = 0) or (h.journaltype=1 and h.docdate < @startdate ))

		` + closeDocFilter + `
		)
		, prd as (
			select accountcode, sum(debitamount) as debitamount, sum(creditamount) as creditamount
			from journal_doc where journal_doc.docdate between @startdate and @enddate
			group by accountcode
		)
		, journal_sheet_sum as (
			select chart.holdingcode, chart.parid
            , chart.accountcode, chart.accountname
			, chart.accountcategory, chart.accountbalancetype, chart.accountgroup, chart.accountlevel, chart.consolidateaccountcode
			, coalesce(prd.debitamount, 0) as debitamount, coalesce(prd.creditamount, 0) as creditamount
			, case when(accountcategory = 5) then coalesce(prd.debitamount, 0)-coalesce(prd.creditamount, 0)
				else coalesce(prd.creditamount, 0)-coalesce(prd.debitamount, 0)
				end as amount
			from chartofaccounts as chart
			left join prd on prd.accountcode = chart.accountcode
			where chart.holdingcode= @holdingcode and chart.accountcategory in (4,5)
		)
		select holdingcode, parid
        , accountcode, accountname, accountcategory, accountbalancetype
        , accountgroup, accountlevel, consolidateaccountcode
		, debitamount, creditamount, amount
		from journal_sheet_sum
		where amount <> 0
		order by accountcode
	`

	var details []models.ProfitAndLossSheetAccountDetail

	condition := map[string]interface{}{
		"holdingcode": holdingCode,
		"start_date":  startDate,
		"end_date":    endDate,
	}

	if len(accountGroup) > 0 {
		condition["accountgroup"] = accountGroup
	}

	_, err := repo.pst.Raw(query, condition, &details)
	if err != nil {
		return nil, err
	}

	return details, nil
}

func (repo JournalReportPgRepository) GetDataBalanceSheet(holdingCode string, accountGroup string, includeCloseAccountMode bool, endDate time.Time) ([]models.BalanceSheetAccountDetail, error) {

	reportYear := endDate.Year()
	var closeDocFilter string

	accountGroupFilter := ``
	if len(accountGroup) > 0 {
		accountGroupFilter = ` and h.accountgroup = @accountgroup `
	}

	if includeCloseAccountMode == true {
		closeDocFilter = `
		union all
		select h.holdingcode, h.docno, h.docdate, h.accountyear, h.accountperiod, h.accountgroup
			, d.accountcode
			, d.debitamount ,d.creditamount
			, acc.accountcategory
			from journals_detail as d
			join journals as h on h.holdingcode = d.holdingcode and h.docno = d.docno
			left join chartofaccounts as acc on acc.holdingcode = d.holdingcode and acc.accountcode = d.accountcode

            where h.holdingcode= @holdingcode and ( h.journaltype = 1 and (extract (year from h.docdate)) = @reportyear )
		`
	}

	/*

		and (( h.journaltype = 0) or (h.journaltype=1 and date(h.docdate) < date(@enddate) ))

				and (
					(acc.accountcategory in (1,2,3)) or
					(
						acc.accountcategory in (4,5) and (extract (year from h.docdate)) = @reportyear
					)
				)
	*/

	query := `
	WITH journal_doc as (
		select h.holdingcode, h.docno, h.docdate, h.accountyear, h.accountperiod, h.accountgroup
			, d.accountcode
			, d.debitamount ,d.creditamount
			, acc.accountcategory
			from journals_detail as d
			join journals as h on h.holdingcode = d.holdingcode and h.docno = d.docno
			left join chartofaccounts as acc on acc.holdingcode = d.holdingcode and acc.accountcode = d.accountcode

			where h.holdingcode= @holdingcode ` + accountGroupFilter + `  and h.docdate < @enddate
			and (
				acc.accountcategory in (1,2,3) or (acc.accountcategory in (4,5) and (extract (year from h.docdate)) = @reportyear)
			)
			and ( (h.journaltype = 0) or (
				h.journaltype=1 and (extract (year from h.docdate)) < @reportyear
			))

		` + closeDocFilter + `
		)
		, nex as (
			select accountcode, sum(debitamount) as debitamount, sum(creditamount) as creditamount
			from journal_doc where journal_doc.docdate <= @enddate
			group by accountcode
		)
		, journal_sheet_sum as (
			select chart.holdingcode, chart.parid
            , chart.accountcode, chart.accountname
			, chart.accountcategory, chart.accountbalancetype, chart.accountgroup, chart.accountlevel, chart.consolidateaccountcode
			, case when(accountcategory = 1 or accountcategory = 5) then coalesce(nex.debitamount, 0)-coalesce(nex.creditamount, 0)
				else coalesce(nex.creditamount, 0)-coalesce(nex.debitamount, 0)
				end as amount
			from chartofaccounts as chart
			left join nex on nex.accountcode = chart.accountcode
			where chart.holdingcode= @holdingcode
		)
		select holdingcode, parid
        , accountcode, accountname, accountcategory, accountbalancetype
        , accountgroup, accountlevel, consolidateaccountcode
		, amount
		from journal_sheet_sum
		where amount <> 0
		order by accountcode
	`

	var details []models.BalanceSheetAccountDetail

	condition := map[string]interface{}{
		"holdingcode": holdingCode,
		"end_date":    endDate,
		"reportyear":  reportYear,
	}

	if len(accountGroup) > 0 {
		condition["accountgroup"] = accountGroup
	}

	_, err := repo.pst.Raw(query, condition, &details)
	if err != nil {
		return nil, err
	}

	fmt.Print("query : " + query + "\n")
	fmt.Printf("details: \n %+v\n", details)

	return details, nil
}

func (repo JournalReportPgRepository) GetDataLedgerAccount(
	holdingCode string,
	accountGroup string,
	creditorCode string,
	debtorCode string,
	consolidateAccountCode string,
	accountRanges []models.LedgerAccountCodeRange,
	bookCode string,
	startDate time.Time,
	endDate time.Time,
) ([]models.LedgerAccountRaw, error) {

	accountCodeQuery := ""
	accountGroupQuery := ""
	consolidateAccountCodeQuery := ""

	creditorQuery := ""
	debtorQuery := ""

	values := map[string]interface{}{
		"holdingcode": holdingCode,
		"start_date":  startDate,
		"end_date":    endDate,
	}

	if len(accountRanges) > 0 {

		for idx, accRange := range accountRanges {
			if len(accountCodeQuery) > 0 {
				accountCodeQuery = accountCodeQuery + " or "
			}

			idxStr := strconv.Itoa(idx)
			accountCodeQuery = accountCodeQuery + " a.accountcode between @accountcode" + idxStr + "1 and  @accountcode" + idxStr + "2"

			values["accountcode"+idxStr+"1"] = accRange.Start
			values["accountcode"+idxStr+"2"] = accRange.End

		}
	}

	if len(accountGroup) > 0 {
		accountGroupQuery = " AND a.accountgroup = @accountgroup"
		values["accountgroup"] = accountGroup
	}

	if len(consolidateAccountCode) > 0 {
		consolidateAccountCodeQuery = " AND a.consolidateaccountcode = @consolidateaccountcode"
		values["consolidateaccountcode"] = consolidateAccountCode
	}

	if len(accountCodeQuery) > 0 {
		accountCodeQuery = " and ( " + accountCodeQuery + " ) "
	}

	if len(creditorCode) > 0 {
		creditorQuery = " and j.creditor->>'code' = @creditorcode"
		values["creditorcode"] = creditorCode
	}

	if len(debtorCode) > 0 {
		debtorQuery = " and j.debtor->>'code' = @debtorcode"
		values["debtorcode"] = debtorCode
	}

	if len(bookCode) > 0 {
		creditorQuery = creditorQuery + " and j.book_code = @bookcode"
		values["bookcode"] = bookCode
	}

	rawQuery := `select * from (
		WITH
			acc as (
			SELECT  a.accountcode,a.accountname,a.accountgroup, a.consolidateaccountcode
		from chartofaccounts a  WHERE holdingcode = @holdingcode ` + accountGroupQuery + consolidateAccountCodeQuery + accountCodeQuery + `
		)
		,
		acc_balance as (
		select d.accountcode,  sum(d.debitamount -  d.creditamount) as amount
		from  journals j
                left join journals_detail d on j.holdingcode = d.holdingcode AND j.docno = d.docno
                left join chartofaccounts a on j.holdingcode = a.holdingcode AND d.accountcode = a.accountcode
		where j.docdate < @startdate and j.holdingcode = @holdingcode ` + accountGroupQuery + creditorQuery + debtorQuery + consolidateAccountCodeQuery + accountCodeQuery + `
		group by d.accountcode
		)
		SELECT -1 as rowmode, '1900-01-01'::date as docdate, '' as docno,acc.accountcode,acc.accountname, '' as accountdescription,
		0 as debitamount, 0 as creditamount, COALESCE(amount, 0) as amount, acc.accountgroup, acc.consolidateaccountcode
		FROM acc left join acc_balance ON acc.accountcode = acc_balance.accountcode
		union all
		select 0 as rowmode, j.docdate, j.docno, d.accountcode,d.accountname, j.accountdescription as accountdescription, d.debitamount, d.creditamount, 0 as amount,a.accountgroup, a.consolidateaccountcode
		from journals j
		join journals_detail d on j.holdingcode = d.holdingcode and j.docno = d.docno
		join chartofaccounts a on a.holdingcode = j.holdingcode and a.accountcode = d.accountcode
		where j.docdate between @startdate and @enddate and j.holdingcode = @holdingcode ` + accountGroupQuery + creditorQuery + debtorQuery + consolidateAccountCodeQuery + accountCodeQuery + `
			) as final_data order by accountcode,rowmode,docdate`

	rawDocList := []models.LedgerAccountRaw{}

	_, err := repo.pst.Raw(rawQuery, values, &rawDocList)
	if err != nil {
		return nil, err
	}

	return rawDocList, nil

}

func (repo JournalReportPgRepository) GetMultiShopRevenue(
	holdingCodes []string,
	startDate time.Time,
	endDate time.Time,
) ([]models.MultiShopRevenueRaw, error) {

	query := `
	WITH journal_doc AS (
		SELECT
			h.holdingcode,
			h.docno,
			h.docdate,
			d.accountcode,
			d.debitamount,
			d.creditamount,
			acc.accountcategory
		FROM journals_detail AS d
		JOIN journals AS h
			ON h.holdingcode = d.holdingcode AND h.docno = d.docno
		LEFT JOIN chartofaccounts AS acc
			ON acc.holdingcode = d.holdingcode AND acc.accountcode = d.accountcode
		WHERE
			h.holdingcode = ANY(@holdingcodes)
			AND h.docdate BETWEEN @startdate AND @enddate
			AND h.journaltype = 0
			AND acc.accountcategory IN (4, 5)
	),
	aggregated AS (
		SELECT
			holdingcode,
			accountcategory,
			SUM(creditamount - debitamount) AS totalamount
		FROM journal_doc
		GROUP BY holdingcode, accountcategory
	)
	SELECT
		holdingcode,
		accountcategory,
		totalamount
	FROM aggregated
	ORDER BY holdingcode, accountcategory
	`

	var results []models.MultiShopRevenueRaw

	condition := map[string]interface{}{
		"holdingcodes": pq.Array(holdingCodes),
		"start_date":   startDate,
		"end_date":     endDate,
	}

	_, err := repo.pst.Raw(query, condition, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}
