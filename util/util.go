package util

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
)

func FetchData(url, traceID, parentID string, ctx context.Context) ([]byte, error) {

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
