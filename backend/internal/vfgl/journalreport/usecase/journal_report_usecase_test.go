package usecase_test

import (
	"smlcloudplatform/internal/vfgl/journalreport/usecase"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJournalSheetReportBalanceSideUsecase(t *testing.T) {

	usecase := &usecase.TrialBalanceSheetReportUsecase{}

	type args struct {
		accType int16
		amount  float64
	}

	cases := []struct {
		name     string
		args     args
		wantErr  bool
		wantData bool
	}{
		{
			name: "Account Mode Assert Should be credit side",
			args: args{
				accType: 1,
				amount:  200,
			},
			wantData: true,
		},
		{
			name: "Account Mode Assert And Lower Than zero Should be credit side",
			args: args{
				accType: 1,
				amount:  -200,
			},
			wantData: false,
		},
		{
			name: "Account Mode Cost Should be credit side",
			args: args{
				accType: 3,
				amount:  200,
			},
			wantData: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := usecase.IsAmountDebitSide(tt.args.accType, tt.args.amount)
			assert.Equal(t, tt.wantData, got)
		})

	}
}
