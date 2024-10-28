package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/egoavara/temporal-for-crosstx/event"
	"go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func main() {

	tclt, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer tclt.Close()

	conn, err := pgx.Connect(context.Background(), "postgres://temporal:temporal@localhost:5432/postgres")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	w := worker.New(tclt, "temporal-for-crosstx-taskqueue", worker.Options{
		BackgroundActivityContext: context.WithValue(
			context.Background(),
			event.CtxPgxConn, conn,
		),
	})

	w.RegisterWorkflow(event.CrossTx)
	w.RegisterActivity(event.CommitTx)
	go func() {
		g := gin.Default()
		g.GET("/data/:title", func(c *gin.Context) {
			wf := must(tclt.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
				TaskQueue: "temporal-for-crosstx-taskqueue",
			}, event.CrossTx, c.Param("title")))
			c.JSON(200, gin.H{"workflow_id": wf.GetID()})
		})
		g.GET("/data/:title/update", func(c *gin.Context) {
			wf := GetWorkflowByTitle(tclt, c.Param("title"))
			up := must(tclt.UpdateWorkflow(context.Background(), client.UpdateWorkflowOptions{
				WorkflowID: wf.GetExecution().WorkflowId,
				UpdateName: "patch",
				Args: []interface{}{
					string(must(json.Marshal(map[string]any{
						"contents": c.Query("patch"),
					}))),
				},
				WaitForStage: client.WorkflowUpdateStageCompleted,
			}))
			var out event.Data
			if err := up.Get(context.Background(), &out); err != nil {
				panic(err)
			}
			c.JSON(200, out)
		})
		g.GET("/data/:title/current", func(c *gin.Context) {
			wf := GetWorkflowByTitle(tclt, c.Param("title"))
			fmt.Println(wf.GetExecution().WorkflowId)
			qr := must(tclt.QueryWorkflow(context.Background(), wf.GetExecution().WorkflowId, "", "get"))
			var out event.Data
			if err := qr.Get(&out); err != nil {
				panic(err)
			}
			c.JSON(200, out)
		})
		g.GET("/data/:title/commit", func(c *gin.Context) {
			wf := GetWorkflowByTitle(tclt, c.Param("title"))
			up := must(tclt.UpdateWorkflow(context.Background(), client.UpdateWorkflowOptions{
				WorkflowID:   wf.GetExecution().WorkflowId,
				UpdateName:   "commit",
				Args:         []any{},
				WaitForStage: client.WorkflowUpdateStageCompleted,
			}))
			var out event.Data
			if err := up.Get(context.Background(), &out); err != nil {
				panic(err)
			}
			c.JSON(200, out)
		})
		g.Run("0.0.0.0:9000")
	}()
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

func GetWorkflowByTitle(tclt client.Client, title string) *workflow.WorkflowExecutionInfo {
	wfs, err := tclt.ListWorkflow(context.Background(), &workflowservice.ListWorkflowExecutionsRequest{
		Query: fmt.Sprintf("Title = '%s'", title),
	})
	if err != nil {
		panic(err)
	}
	if len(wfs.Executions) == 0 {
		panic("No workflow found")
	}
	return wfs.Executions[0]
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}
