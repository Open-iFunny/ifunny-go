package ifunny

import (
	"encoding/json"
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

type authScheme string

const (
	BEARER = authScheme("bearer")
	BASIC  = authScheme("basic")
)

func MakeClient(scheme authScheme, token, userAgent string) (*Client, error) {
	client := &Client{scheme, token, userAgent, http.DefaultClient, logrus.New(), nil}
	client.log.SetFormatter(&logrus.JSONFormatter{})
	client.log.SetLevel(LogLevel)

	if scheme == BEARER {
		self, err := client.GetUser(compose.UserAccount())
		if err != nil {
			return nil, err
		}

		client.Self = self
	}

	return client, nil
}

func MakeClientLog(scheme authScheme, token, userAgent string, log *logrus.Logger) (*Client, error) {
	client, err := MakeClient(scheme, token, userAgent)
	if err != nil {
		return nil, err
	}

	client.log = log
	return client, nil
}

type Client struct {
	scheme           authScheme
	token, userAgent string
	http             *http.Client
	log              *logrus.Logger

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
			return nil, fmt.Errorf("failed to unwrap HTTP error: %s", err)
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
		"authorization":     []string{string(client.scheme) + " " + client.token},
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
