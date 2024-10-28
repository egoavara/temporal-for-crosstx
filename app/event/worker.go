package event

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tidwall/gjson"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type CtxKey string

const (
	CtxPgxConn CtxKey = "CtxPgxConn"
)

var (
	SearchAttributeTitle = temporal.NewSearchAttributeKeyKeyword("Title")
)

type Data struct {
	Id       int64
	Title    string
	Contents json.RawMessage
}

func mustdo(err error) {
	if err != nil {
		panic(err)
	}
}
func CrossTx(ctx workflow.Context, title string) (*Data, error) {

	var logger = workflow.GetLogger(ctx)
	var data = &Data{
		Title: title,
	}
	var isCommitted bool = false

	logger.Info("Run workflow.", "Title", title)
	mustdo(workflow.UpsertTypedSearchAttributes(ctx, SearchAttributeTitle.ValueSet(title)))
	mustdo(workflow.SetQueryHandler(ctx, "get", func() (*Data, error) {
		return data, nil
	}))
	mustdo(workflow.SetUpdateHandler(ctx, "patch", func(ctx workflow.Context, patch string) (*Data, error) {
		title := gjson.Get(patch, "title")
		if title.Exists() {
			data.Title = title.String()
			mustdo(workflow.UpsertTypedSearchAttributes(ctx, SearchAttributeTitle.ValueSet(data.Title)))
		}
		contents := gjson.Get(patch, "contents")
		if contents.Exists() {
			data.Contents = json.RawMessage(contents.Raw)
		}
		return data, nil
	}))

	mustdo(workflow.SetUpdateHandler(ctx, "commit", func(ctx workflow.Context) (*Data, error) {
		if isCommitted {
			return data, nil
		}
		ctx = workflow.WithLocalActivityOptions(ctx, workflow.LocalActivityOptions{
			ScheduleToCloseTimeout: 10 * time.Second,
		})
		err := workflow.ExecuteLocalActivity(ctx, CommitTx, *data).
			Get(ctx, &data)
		if err != nil {
			return nil, err
		}
		isCommitted = true
		return data, nil
	}))

	mustdo(workflow.Await(ctx, func() bool { return isCommitted }))
	return data, nil
}

func CommitTx(ctx context.Context, data Data) (*Data, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Data got.", "data", data)
	conn := ctx.Value(CtxPgxConn).(*pgx.Conn)
	logger.Info("Data conn.", "conn", conn)
	row := conn.QueryRow(ctx, "INSERT INTO data (title, contents) VALUES ($1, $2) RETURNING id, title, contents", data.Title, data.Contents)
	var out Data
	err := row.Scan(&out.Id, &out.Title, &out.Contents)
	if err != nil {
		logger.Error("Data commit failed.", "err", err)
		return nil, err
	}
	logger.Info("Data committed.", "Id", out.Id)
	return &out, nil
}
