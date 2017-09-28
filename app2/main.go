package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-xray-sdk-go/xray"
)

type Github struct {
	UserURL string `json:"user_url"`
}

func main() {
	xray.Configure(xray.Config{LogLevel: "info"})

	http.Handle("/", xray.Handler(xray.NewFixedSegmentNamer("app2"), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var traceID, parentID string
		if seg, ok := r.Context().Value(xray.ContextKey).(*xray.Segment); ok {
			traceID = seg.TraceID
			parentID = seg.ID
		}

		ctx, subseg := xray.BeginSegment(r.Context(), "GithubCall")
		subseg.TraceID, subseg.ParentID = traceID, parentID

		githubResp, err := fetchData("https://api.github.com", traceID, parentID, ctx)
		if err != nil {
			subseg.Close(err)
			http.Error(w, err.Error(), 500)
			return
		}

		var githubData Github
		json.Unmarshal(githubResp, &githubData)
		subseg.Close(nil)

		databaseData, err := fetchData("http://app3:8083", traceID, parentID, ctx)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		data := map[string]map[string]string{
			"app2": map[string]string{
				"github_user_url": githubData.UserURL,
				"database_data":   string(databaseData),
			},
		}

		output, err := json.Marshal(data)

		w.Write([]byte(output))
	})))

	http.ListenAndServe(":8082", nil)
}

func fetchData(url, traceID, parentID string, ctx context.Context) ([]byte, error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Header.Add("x-amzn-trace-id", fmt.Sprintf("Sampled=1;Root=%s;Parent=%s", traceID, parentID))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	return body, err

}
