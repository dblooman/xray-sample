package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-xray-sdk-go/xray"
	_ "github.com/lib/pq"
)

func main() {
	xray.Configure(xray.Config{LogLevel: "info"})

	http.Handle("/", xray.Handler(xray.NewFixedSegmentNamer("app3"), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var traceID, parentID string
		if seg, ok := r.Context().Value(xray.ContextKey).(*xray.Segment); ok {
			traceID = seg.TraceID
			parentID = seg.ID
		}

		ctx, seg := newSubSegmnet(r.Context(), traceID, parentID, "app3-db-call")

		row, err := queryDB(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		seg.Close(nil)

		output, _ := json.Marshal(map[string]sql.Result{"app3": row})

		w.Write(output)
	})))

	http.ListenAndServe(":8083", nil)
}

func newSubSegmnet(ctx context.Context, traceID, parentID, name string) (context.Context, *xray.Segment) {
	ctx, seg := xray.BeginSubsegment(ctx, name)
	seg.TraceID = traceID
	seg.ParentID = parentID

	return ctx, seg
}

func queryDB(ctx context.Context) (sql.Result, error) {
	db, err := xray.SQL("postgres", "postgres://postgres:postgres@db:5432/example?sslmode=disable")
	if err != nil {
		return nil, err
	}
	resp, err := db.Exec(ctx, "SELECT 1")
	if err != nil {
		return nil, err
	}

	return resp, nil
}
