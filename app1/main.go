package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/DaveBlooman/xray-sample/util"
	"github.com/aws/aws-xray-sdk-go/xray"
)

type AppData struct {
	App2 struct {
		DatabaseData  string `json:"database_data"`
		GithubUserURL string `json:"github_user_url"`
	} `json:"app2"`
}

func main() {
	xray.Configure(xray.Config{LogLevel: "info"})

	http.Handle("/", xray.Handler(xray.NewFixedSegmentNamer("app1"), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// This will give us the initial ID for the top level Segment as well as the TraceID.
		var traceID, parentID string
		if seg, ok := r.Context().Value(xray.ContextKey).(*xray.Segment); ok {
			traceID = seg.TraceID
			parentID = seg.ID
		}

		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()

		appdata, err := util.FetchData("http://app2:8082", traceID, parentID, ctx)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var data AppData
		json.Unmarshal(appdata, &data)

		finalOutput := map[string]AppData{
			"app1": data,
		}

		output, err := json.Marshal(finalOutput)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Write([]byte(output))
	})))

	http.ListenAndServe(":8081", nil)
}
