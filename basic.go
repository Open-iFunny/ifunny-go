package ifunny

import (
	"context"
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

	// APP_VERSION and APP_BUILD identify the iFunny mobile app release
	// we impersonate in the user-agent. Pinned to values from
	// makeshiftartist/ifunny.ts. These will go stale as iFunny ships
	// new builds; treat as a maintenance point rather than a stable API.
	APP_VERSION = "8.15.1"
	APP_BUILD   = "1130736"
)

// UserAgent is anything that renders itself to an iFunny user-agent
// string. Both device profiles (Android, IOS) and caller-supplied raw
// strings (RawUserAgent) satisfy this. Client constructors accept a
// UserAgent so callers can either use the built-in device abstraction
// or supply their own.
type UserAgent interface {
	String() string
}

// RawUserAgent is a caller-supplied user-agent string that satisfies
// the UserAgent interface unchanged. Use this when you want full
// control over the UA (e.g. reading it from configuration or an
// environment variable).
type RawUserAgent string

// String returns the raw user-agent string.
func (r RawUserAgent) String() string { return string(r) }

// Phone models a mobile device the client pretends to be. iFunny is a
// mobile app whose backend gates behavior on the user-agent, so any
// caller-generated UA needs to look like a real phone. Implementations
// own their OS name, OS version, brand, and model strings, and render
// themselves into a UserAgent via UserAgent().
type Phone interface {
	OS() string        // "Android", "iOS"
	OSVersion() string // caller-supplied, e.g. "14", "17.5.1"
	Brand() string     // e.g. "google", "Apple"
	Model() string     // e.g. "Pixel 8", "iPhone 15 Pro"

	UserAgent() UserAgent
}

// Android models a Google Pixel 8 running the caller-supplied Android
// version. Version is a free-form string (e.g. "14", "15") so new
// releases don't require an SDK update.
type Android struct{ Version string }

// OS returns "Android".
func (a Android) OS() string { return "Android" }

// OSVersion returns the Android version supplied by the caller.
func (a Android) OSVersion() string { return a.Version }

// Brand returns "google".
func (a Android) Brand() string { return "google" }

// Model returns "Pixel 8".
func (a Android) Model() string { return "Pixel 8" }

// UserAgent returns a user-agent string for this Android device.
func (a Android) UserAgent() UserAgent { return renderPhoneUA(a) }

// IOS models an iPhone 15 Pro running the caller-supplied iOS version.
// Version is a free-form string (e.g. "17.5.1", "18.0") so new releases
// don't require an SDK update.
type IOS struct{ Version string }

// OS returns "iOS".
func (i IOS) OS() string { return "iOS" }

// OSVersion returns the iOS version supplied by the caller.
func (i IOS) OSVersion() string { return i.Version }

// Brand returns "Apple".
func (i IOS) Brand() string { return "Apple" }

// Model returns "iPhone 15 Pro".
func (i IOS) Model() string { return "iPhone 15 Pro" }

// UserAgent returns a user-agent string for this iOS device.
func (i IOS) UserAgent() UserAgent { return renderPhoneUA(i) }

// renderPhoneUA renders a Phone into the iFunny app's UA template:
//
//	iFunny/{APP_VERSION}({APP_BUILD}) {OS}/{OSVersion} ({Brand}; {Model}; {Brand})
func renderPhoneUA(p Phone) UserAgent {
	return RawUserAgent(fmt.Sprintf("iFunny/%s(%s) %s/%s (%s; %s; %s)",
		APP_VERSION, APP_BUILD,
		p.OS(), p.OSVersion(),
		p.Brand(), p.Model(), p.Brand(),
	))
}

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

// PrimeBasic activates the basic token this client was constructed with: one
// GET /counters, then the server-side ~15s wait. Call once on a freshly
// generated token before making other requests. Uses the client's configured
// *http.Client, so any custom transport/timeouts are honored. The ctx governs
// both the /counters request and the subsequent wait, so a cancelled ctx aborts
// priming promptly and returns ctx.Err().
func (client *Client) PrimeBasic(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", client.apiRoot+"/counters", nil)
	if err != nil {
		return err
	}
	req.Header = client.header()

	resp, err := client.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("prime basic token: HTTP %d", resp.StatusCode)
	}

	select {
	case <-time.After(15 * time.Second):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
