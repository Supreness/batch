package example2

import (
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/supreness/batch"
	"github.com/supreness/batch/util"
	"log"
	"time"
)

func openDB() *sql.DB {
	var sqlDb *sql.DB
	var err error
	sqlDb, err = sql.Open("mysql", "root:root123@tcp(127.0.0.1:3306)/example?charset=utf8&parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	return sqlDb
}

func removeJobData() {
	sqlDb := openDB()
	_, err := sqlDb.Exec("DELETE FROM t_trade")
	if err != nil {
		log.Fatal(err)
	}
	_, err = sqlDb.Exec("DELETE FROM t_repay_plan")
	if err != nil {
		log.Fatal(err)
	}
}

func buildAndRunJob() {
	sqlDb := openDB()
	batch.SetDB(sqlDb)
	batch.SetTransactionManager(batch.NewTransactionManager(sqlDb))

	step1 := batch.NewStep("import_trade").ReadFile(tradeFile).Writer(&tradeImporter{sqlDb}).Partitions(10).Build()
	step2 := batch.NewStep("gen_repay_plan").Reader(&tradeReader{sqlDb}).Handler(&repayPlanHandler{sqlDb}).Partitions(10).Build()
	step3 := batch.NewStep("stats").Handler(&statsHandler{sqlDb}).Build()
	step4 := batch.NewStep("export_trade").Reader(&tradeReader{sqlDb}).WriteFile(tradeFileExport).Partitions(10).Build()
	step5 := batch.NewStep("upload_file_to_ftp").CopyFile(copyFileToFtp, copyChecksumFileToFtp).Build()
	job := batch.NewJob("accounting_job").Step(step1, step2, step3, step4, step5).Build()

	batch.Register(job)

	params, _ := util.JsonString(map[string]interface{}{
		"date": time.Now().Format("2006-01-02"),
		"rand": time.Now().Nanosecond(),
	})
	batch.Start(context.Background(), job.Name(), params)

}
