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

func MakeClient(bearer, userAgent string) (*Client, error) {
	client := &Client{bearer, userAgent, http.DefaultClient, logrus.New(), nil}
	client.log.SetFormatter(&logrus.JSONFormatter{})
	client.log.SetLevel(LogLevel)

	self, err := client.GetUser(compose.UserAccount())
	if err != nil {
		return nil, err
	}

	client.Self = self
	return client, nil
}

func MakeClientLog(bearer, userAgent string, log *logrus.Logger) (*Client, error) {
	client, err := MakeClient(bearer, userAgent)
	if err != nil {
		return nil, err
	}

	client.log = log
	return client, nil
}

type Client struct {
	bearer, userAgent string
	http              *http.Client
	log               *logrus.Logger

	Self *User
}

func request(desc compose.Request, header http.Header, client *http.Client) (*http.Response, error) {
	request, err := http.NewRequest(desc.Method, apiRoot+desc.Path, desc.Body)
	if err != nil {
		return nil, err
	}
	request.Header = header
	request.URL.RawQuery = desc.Query.Encode()

	return client.Do(request)
}

func (client *Client) header() http.Header {
	return http.Header{
		"authorization":     []string{"bearer " + client.bearer},
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
