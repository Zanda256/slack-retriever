package slack

import (
	"context"
	"crypto/x509"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/time/rate"
)

const (
	slackAPIURL   = "https://slack.com/api"
	convoInfo     = "conversations.info"
	convoMembers  = "conversations.members"
	convoMessages = "conversations.history"
)

var (
	//Tier4Rl Limits the rate of requests to 100 per minute
	Tier4Rl = rate.NewLimiter(rate.Every(1*time.Minute), 100) // 50 request every 1 minute
	//Tier3Rl Limits the rate of requests to 50 per minute
	Tier3Rl = rate.NewLimiter(rate.Every(1*time.Minute), 50) // 50 request every 1 minute

	// A regular expression to match the error returned by net/http when the
	// configured number of redirects is exhausted. This error isn't typed
	// specifically so we resort to matching on the error string.
	redirectsErrorRe = regexp.MustCompile(`stopped after \d+ redirects\z`)

	// A regular expression to match the error returned by net/http when the
	// scheme specified in the URL is invalid. This error isn't typed
	// specifically so we resort to matching on the error string.
	schemeErrorRe = regexp.MustCompile(`unsupported protocol scheme`)

	//MyRetryPolicy specifies conditions under which the http.Client should retry a failed request. Returns true if request should be retried.
	MyRetryPolicy = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		if err != nil {
			if v, ok := err.(*url.Error); ok {
				// Don't retry if the error was due to too many redirects.
				if redirectsErrorRe.MatchString(v.Error()) {
					return false, v
				}

				// Don't retry if the error was due to an invalid protocol scheme.
				if schemeErrorRe.MatchString(v.Error()) {
					return false, v
				}

				// Don't retry if the error was due to TLS cert verification failure.
				if _, ok := v.Err.(x509.UnknownAuthorityError); ok {
					return false, v
				}
			}

			// The error is likely recoverable so retry.
			return true, nil
		}
		retryAfterCodes := []int{413, 429, 503, 408, 423, 504}
		contains := func(s []int, str int) bool {
			for _, v := range s {
				if v == str {
					return true
				}
			}
			return false
		}
		if contains(retryAfterCodes, resp.StatusCode) {
			return true, nil
		}
		return false, nil
	}
)

//Client with methods to make http requests to the slack api
type Client struct {
	HTTPcl *http.Client
	RateLm *rate.Limiter
}

//Create a new http.Client for the slack client
//Adds retry settings to the new http.Client
func newHTTPcl() *http.Client {
	retryClient := retryablehttp.NewClient()

	retryClient.RetryWaitMin = (1 * time.Second)
	retryClient.RetryWaitMax = (2 * time.Second)
	retryClient.RetryMax = 5

	retryClient.CheckRetry = retryablehttp.CheckRetry(MyRetryPolicy)
	//convert the retryablehttp.Client into standard http.Client but with
	//our custom retry settings
	return retryClient.StandardClient()
}

//NewClient creates a new http.Client configured with custom retry and rate limiting settings
func NewClient() *Client {
	clnt := newHTTPcl()
	sc := &Client{
		HTTPcl: clnt,
		RateLm: Tier3Rl,
	}
	return sc
}

//DoRequest performs http rate limited requests
func (c *Client) DoRequest(req *http.Request) (*http.Response, error) {
	// Comment out the below 5 lines to turn off ratelimiting
	ctx := context.Background()
	err := c.RateLm.Wait(ctx) // This is a blocking call. Honors the rate limit
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPcl.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
