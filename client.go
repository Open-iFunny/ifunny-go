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
// requests. When passed multiple times the last call wins.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) {
		c.http = h
	}
}

// WithLogger sets the logrus.Logger used by the client. When not supplied
// the client uses a logrus.New() logger configured with a JSON formatter at
// LogLevel. When passed multiple times the last call wins.
func WithLogger(log *logrus.Logger) Option {
	return func(c *Client) {
		c.log = log
	}
}

func newClient(authorization string, ua UserAgent, opts ...Option) *Client {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(LogLevel)
	c := &Client{
		userAgent:     ua.String(),
		authorization: authorization,
		http:          http.DefaultClient,
		log:           log,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// MakeClient constructs an authenticated client using a bearer token. It fetches
// the authenticated user's account data and stores it in the returned client's
// Self field. Returns an error if authentication fails or the account fetch fails.
func MakeClient(bearer string, ua UserAgent, opts ...Option) (*Client, error) {
	client := newClient("bearer "+bearer, ua, opts...)
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
func MakeClientBasic(basic string, ua UserAgent, opts ...Option) (*Client, error) {
	return newClient("Basic "+basic, ua, opts...), nil
}

// Client is an authenticated iFunny API client. It holds authentication state
// (bearer token or basic auth), user-agent information, and the authenticated user's
// account data (Self). Use MakeClient or MakeClientBasic to construct; pass
// WithLogger or WithHTTPClient to override defaults.
type Client struct {
	bearer, userAgent string
	authorization     string
	http              *http.Client
	log               *logrus.Logger

	Self *User
}

// APIError represents an HTTP error returned by the iFunny API. It is returned
// wrapped in an error by request methods; use AsAPIError to unwrap it.
type APIError struct {
	Kind        string `json:"error"`
	Description string `json:"error_description"`
	Status      int    `json:"status"`
}

// Error returns a human-readable error message for the API error.
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

// RequestJSON executes a composed API request and unmarshals the response body
// into output as JSON. Returns errors from the network request or JSON decoding.
// API errors (HTTP >= 400) are returned as *APIError wrapped in error.
func (client *Client) RequestJSON(desc compose.Request, output any) error {
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
