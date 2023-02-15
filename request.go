package icloudgo

import (
	"fmt"
	"io"
)

type rawReq struct {
	Method       string
	URL          string
	Headers      map[string]string
	Querys       map[string]string
	Body         any
	To           any
	ExpectStatus set[int]
	Stream       bool
}

func (r *Client) request(req *rawReq) (string, error) {
	text, _, err := r.doRequest(req)
	return text, err
}

func (r *Client) requestStream(req *rawReq) (io.ReadCloser, error) {
	req.Stream = true
	_, body, err := r.doRequest(req)
	return body, err
}

func (r *Client) doRequest(req *rawReq) (string, io.ReadCloser, error) {
	// fmt.Printf("start %s %s\n", req.Method, req.URL)
	status := 0
	// defer func() {
	// 	fmt.Printf("end %s %s status=%d\n", req.Method, req.URL, status)
	// }()

	res := r.httpCli.New(req.Method, req.URL).WithURLCookie("https://icloud.com.cn")
	if len(req.Headers) > 0 {
		res = res.WithHeaders(req.Headers)
	}
	if len(req.Querys) > 0 {
		res = res.WithQuerys(req.Querys)
	}
	if req.Body != nil {
		if len(req.Headers) > 0 && req.Headers["Content-Type"] == "text/plain" {
			res = res.WithBody(req.Body)
		} else {
			res = res.WithJSON(req.Body)
		}
	}

	resp, respErr := res.Response()
	if resp != nil {
		for k, callback := range contextHeader {
			if resp.Header.Get(k) != "" {
				callback(r.sessionData, resp.Header.Get(k))
			}
		}
	}

	status = res.MustResponseStatus()

	if req.Stream {
		if respErr != nil {
			return "", nil, fmt.Errorf("%s %s failed, status %d, err: %s", req.Method, req.URL, status, respErr)
		}
		if req.ExpectStatus != nil && req.ExpectStatus.Len() > 0 && !req.ExpectStatus.Has(status) {
			return "", nil, fmt.Errorf("%s %s failed, expect status %v, but got %d", req.Method, req.URL, req.ExpectStatus.String(), status)
		}
		return "", resp.Body, nil
	}

	text, err := res.Text()
	if err != nil {
		return text, nil, fmt.Errorf("%s %s failed, status %d, err: %s, response text: %s", req.Method, req.URL, status, err, text)
	}

	for _, mayIsErr := range []func([]byte) error{mayErr1, mayErr2, mayErr4, mayErr3} {
		if err := mayIsErr([]byte(text)); err != nil {
			return text, nil, fmt.Errorf("%s %s failed, status %d, err: %w", req.Method, req.URL, status, err)
		}
	}

	if req.ExpectStatus != nil && req.ExpectStatus.Len() > 0 && !req.ExpectStatus.Has(status) {
		return text, nil, fmt.Errorf("%s %s failed, expect status %v, but got %d, response text: %s", req.Method, req.URL, req.ExpectStatus.String(), status, text)
	}

	return text, nil, err
}

func (r *Client) getAuthHeaders(overwrite map[string]string) map[string]string { //            "Accept": "*/*",
	headers := map[string]string{
		"Accept":                           "*/*",
		"Content-Type":                     "application/json",
		"X-Apple-OAuth-Client-Id":          "d39ba9916b7251055b22c7f910e2ea796ee65e98b2ddecea8f5dde8d9d1a815d",
		"X-Apple-OAuth-Client-Type":        "firstPartyAuth",
		"X-Apple-OAuth-Redirect-URI":       "https://www.icloud.com",
		"X-Apple-OAuth-Require-Grant-Code": "true",
		"X-Apple-OAuth-Response-Mode":      "web_message",
		"X-Apple-OAuth-Response-Type":      "code",
		"X-Apple-OAuth-State":              r.clientID,
		"X-Apple-Widget-Key":               "d39ba9916b7251055b22c7f910e2ea796ee65e98b2ddecea8f5dde8d9d1a815d",

		"Origin":     r.homeEndpoint,
		"Referer":    fmt.Sprintf("%s/", r.homeEndpoint),
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36",
	}
	for k, v := range overwrite {
		headers[k] = v
	}
	return headers
}

func (r *Client) getCommonHeaders(overwrite map[string]string) map[string]string { //            "Accept": "*/*",
	headers := map[string]string{
		"Origin":     r.homeEndpoint,
		"Referer":    fmt.Sprintf("%s/", r.homeEndpoint),
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36",
	}
	for k, v := range overwrite {
		headers[k] = v
	}
	return headers
}

var contextHeader = map[string]func(d *SessionData, v string){
	"X-Apple-ID-Account-Country": func(d *SessionData, v string) {
		d.AccountCountry = v
	},
	"X-Apple-ID-Session-Id": func(d *SessionData, v string) {
		d.SessionID = v
	},
	"X-Apple-Session-Token": func(d *SessionData, v string) {
		d.SessionToken = v
	},
	"X-Apple-TwoSV-Trust-Token": func(d *SessionData, v string) {
		d.TrustToken = v
	},
	"scnt": func(d *SessionData, v string) {
		d.Scnt = v
	},
}
