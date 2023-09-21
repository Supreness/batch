package example2

import (
	"database/sql"
	"github.com/supreness/batch"
	"time"
)

type repayPlanHandler struct {
	//TradeReader
	db *sql.DB
}

func (h *repayPlanHandler) Process(item interface{}, chunkCtx *batch.ChunkContext) (interface{}, batch.BatchError) {
	trade := item.(*Trade)
	plans := make([]*RepayPlan, 0)
	restPrincipal := trade.Amount
	for i := 1; i <= trade.Terms; i++ {
		principal := restPrincipal / float64(trade.Terms-i+1)
		interest := restPrincipal * trade.InterestRate / 12
		repayPlan := &RepayPlan{
			AccountNo:  trade.AccountNo,
			LoanNo:     trade.TradeNo,
			Term:       i,
			Principal:  principal,
			Interest:   interest,
			InitDate:   time.Now(),
			RepayDate:  time.Now().AddDate(0, 1, 0),
			RepayState: "",
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		}
		plans = append(plans, repayPlan)
		restPrincipal -= principal
	}
	return plans, nil
}

func (h *repayPlanHandler) Write(items []interface{}, chunkCtx *batch.ChunkContext) batch.BatchError {
	for _, item := range items {
		plans := item.([]*RepayPlan)
		for _, plan := range plans {
			_, err := h.db.Exec("INSERT INTO t_repay_plan(account_no, loan_no, term, principal, interest, init_date, repay_date, repay_state) values (?,?,?,?,?,?,?,?)",
				plan.AccountNo, plan.LoanNo, plan.Term, plan.Principal, plan.Interest, plan.InitDate, plan.RepayDate, plan.RepayState)
			if err != nil {
				return batch.NewBatchError(batch.ErrCodeDbFail, "insert t_repay_plan failed", err)
			}
		}
	}
	return nil
}
