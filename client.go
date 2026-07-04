package ifunny

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/open-ifunny/ifunny-go/compose"
	"github.com/sirupsen/logrus"
)

const (
	apiRoot   = "https://api.ifunny.mobi/v4"
	projectID = "iFunny"

	LogLevel = logrus.InfoLevel
)

// Option configures a Client at construction time. Options are applied before
// any network call the constructor makes (e.g. MakeClient's /account fetch),
// so WithHTTPClient also governs the initial login request.
type Option func(*Client)

// WithHTTPClient sets the underlying *http.Client used for all iFunny API
// requests. A nil client is ignored, leaving the default in place.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) {
		if h != nil {
			c.http = h
		}
	}
}

func newClient(authorization, userAgent string, opts ...Option) *Client {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(LogLevel)
	c := &Client{
		userAgent:     userAgent,
		authorization: authorization,
		http:          http.DefaultClient,
		log:           log,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func MakeClient(bearer, userAgent string, opts ...Option) (*Client, error) {
	client := newClient("bearer "+bearer, userAgent, opts...)
	client.bearer = bearer

	self, err := client.GetUser(compose.UserAccount())
	if err != nil {
		return nil, err
	}

	client.Self = self
	return client, nil
}

// MakeClientBasic builds a client that authenticates with a primed basic token
// (Authorization: Basic <basic>). Unlike MakeClient it does not fetch /account
// (a basic token can't), so Self is nil. Chat requires a bearer and does not
// work on a basic client. Call (*Client).PrimeBasic once on a freshly generated
// token before making other requests.
func MakeClientBasic(basic, userAgent string, opts ...Option) (*Client, error) {
	return newClient("Basic "+basic, userAgent, opts...), nil
}

func MakeClientLog(bearer, userAgent string, log *logrus.Logger, opts ...Option) (*Client, error) {
	client, err := MakeClient(bearer, userAgent, opts...)
	if err != nil {
		return nil, err
	}

	client.log = log
	return client, nil
}

type Client struct {
	bearer, userAgent string
	authorization     string
	http              *http.Client
	log               *logrus.Logger

	Self *User
}

type APIError struct {
	Kind        string `json:"error"`
	Description string `json:"error_description"`
	Status      int    `json:"status"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("HTTP %d: %s: %s", e.Status, e.Kind, e.Description)
}

// AsAPIError unwraps err into an *APIError. The client returns API errors
// as *APIError, so use this rather than asserting on the value type.
func AsAPIError(err error) (*APIError, bool) {
	apiErr := new(APIError)
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}

func request(desc compose.Request, header http.Header, client *http.Client) (*http.Response, error) {
	request, err := http.NewRequest(desc.Method, apiRoot+desc.Path, desc.Body)
	if err != nil {
		return nil, err
	}
	request.Header = header
	request.URL.RawQuery = desc.Query.Encode()

	r, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if r.StatusCode >= 400 {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("failed collecting HTTP error: %s", err)
		}

		unwrap := new(struct {
			Msg []byte `json:"msg"`
		})
		if err := json.Unmarshal(b, unwrap); err != nil {
			return nil, fmt.Errorf("failed to unwrap HTTP error: %s, body: %s", err, string(b))
		}

		apiErr := new(APIError)
		if err := json.Unmarshal(b, apiErr); err != nil {
			return nil, fmt.Errorf("failed to decode HTTP error: %s", err)
		}

		return nil, apiErr
	}

	return r, nil
}

func (client *Client) header() http.Header {
	return http.Header{
		"authorization":     []string{client.authorization},
		"user-agent":        []string{client.userAgent},
		"ifunny-project-id": []string{projectID},
	}
}

func (client *Client) RequestJSON(desc compose.Request, output interface{}) error {
	traceID := uuid.New().String()
	log := client.log.WithFields(logrus.Fields{
		"trace_id": traceID,
		"path":     desc.Path,
		"method":   desc.Method,
		"query":    desc.Query.Encode(),
		"has_body": desc.Body != nil},
	)

	log.Trace("make request")
	response, err := request(desc, client.header(), client.http)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Trace(fmt.Sprintf("got response %s", response.Status))
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Trace(fmt.Sprintf("got response body %s", string(bodyBytes)))
	err = json.Unmarshal(bodyBytes, output)
	if err != nil {
		log.Error(err)
	}

	return err
}
