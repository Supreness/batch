package test

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/supreness/batch"
	"github.com/supreness/batch/util"
	"time"
)

// simple task
func mytask() {
	fmt.Println("mytask executed")
}

//reader
type myReader struct {
}

func (r *myReader) Read(chunkCtx *batch.ChunkContext) (interface{}, batch.BatchError) {
	curr, _ := chunkCtx.StepExecution.StepContext.GetInt("read.num", 0)
	if curr < 100 {
		chunkCtx.StepExecution.StepContext.Put("read.num", curr+1)
		return fmt.Sprintf("value-%v", curr), nil
	}
	return nil, nil
}

//processor
type myProcessor struct {
}

func (r *myProcessor) Process(item interface{}, chunkCtx *batch.ChunkContext) (interface{}, batch.BatchError) {
	return fmt.Sprintf("processed-%v", item), nil
}

//writer
type myWriter struct {
}

func (r *myWriter) Write(items []interface{}, chunkCtx *batch.ChunkContext) batch.BatchError {
	fmt.Printf("write: %v\n", items)
	return nil
}

func main() {
	//set db for batch to store job&step execution context
	var db *sql.DB
	var err error
	db, err = sql.Open("mysql", "root:root123@tcp(127.0.0.1:3306)/batch?charset=utf8&parseTime=true")
	if err != nil {
		panic(err)
	}
	batch.SetDB(db)

	//build steps
	step1 := batch.NewStep("mytask").Handler(mytask).Build()
	//step2 := batch.NewStep("my_step").Handler(&myReader{}, &myProcessor{}, &myWriter{}).Build()
	step2 := batch.NewStep("my_step").Reader(&myReader{}).Processor(&myProcessor{}).Writer(&myWriter{}).ChunkSize(10).Build()

	//build job
	job := batch.NewJob("my_job").Step(step1, step2).Build()

	//register job to batch
	batch.Register(job)

	//run
	//batch.StartAsync(context.Background(), job.Name(), "")
	params, _ := util.JsonString(map[string]interface{}{
		"rand": time.Now().Nanosecond(),
	})
	batch.Start(context.Background(), job.Name(), params)
}
