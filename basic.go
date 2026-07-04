package ifunny

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	basicClientID     = "MsOIJ39Q28"
	basicClientSecret = "PTDc3H8a)Vi=UYap"
)

// GenerateBasic mints a fresh, unprimed basic token, mirroring the iFunny
// app's client-side algorithm (the length-112 variant):
//
//	token = base64( HEX + "_" + id + ":" + sha1hex(HEX + ":" + id + ":" + secret) )
//
// where HEX is an uppercased, dash-stripped random UUID.
func GenerateBasic() (string, error) {
	id := strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", ""))
	a := id + "_" + basicClientID + ":"
	b := id + ":" + basicClientID + ":" + basicClientSecret
	sum := sha1.Sum([]byte(b))
	return base64.StdEncoding.EncodeToString([]byte(a + hex.EncodeToString(sum[:]))), nil
}

// PrimeBasic activates a basic token: one GET /counters authenticated with the
// token as Basic auth, then the server-side ~15s wait. Call once before first
// use of a freshly generated token.
func PrimeBasic(basic, userAgent string) error {
	req, err := http.NewRequest("GET", apiRoot+"/counters", nil)
	if err != nil {
		return err
	}
	req.Header.Set("authorization", "Basic "+basic)
	req.Header.Set("user-agent", userAgent)
	req.Header.Set("ifunny-project-id", projectID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("prime basic token: HTTP %d", resp.StatusCode)
	}

	time.Sleep(15 * time.Second)
	return nil
}
