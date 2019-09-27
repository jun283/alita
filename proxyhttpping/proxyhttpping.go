package proxyhttpping

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strconv"
	"time"

	"golang.org/x/net/publicsuffix"
)

// Ping represents HTTP ping request
type Ping struct {
	proxyAddr     string //"10.0.8.202:80"
	Url           string
	Method        string
	Timeout       time.Duration
	RedirectCount int //0: dis-Redirect
	Interval      time.Duration
	TLSSkipVerify bool
	ProxyIP       string
	ProxyPort     string
	ProxyScheme   string //HTTP/HTTPS
	RespPath      string //Redirect path records
}

// Result holds Ping result
type Result struct {
	StatusCode int
	TotalTime  float64
	Size       int
	Proto      string
	Server     string
	Status     string
	Trace      Trace
}

// Trace holds trace results
type Trace struct {
	ConnectionTime  float64
	TimeToFirstByte float64
}

// Normalize fixes scheme
func NormalizeURL(URL string) string {
	re := regexp.MustCompile(`(?i)https{0,1}://`)
	if !re.MatchString(URL) {
		URL = fmt.Sprintf("http://%s", URL)
	}
	return URL
}

func (p *Ping) Client() (*http.Client, error) {
	dialer := &net.Dialer{
		Timeout:   p.Timeout,
		DualStack: true,
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: p.TLSSkipVerify,
		},
		DisableCompression: true,
		DisableKeepAlives:  true,
	}

	tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		//if p.ProxyAddr != "" {
		//	addr = w.proxyAddr
		//}
		//return dialer.DialContext(ctx, network, addr)
		return dialer.Dial(network, p.proxyAddr)
	}

	options := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, err := cookiejar.New(&options)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(p.Timeout),
		Jar:       jar,
	}

	if p.RedirectCount == 0 {
		//DisableRedirects
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	} else {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			p.RespPath += "->" + strconv.Itoa(req.Response.StatusCode) + "->" + req.URL.String()
			if p.RedirectCount >= len(via) {
				return http.ErrUseLastResponse
			}
			return nil
		}
	}

	return client, nil
}
