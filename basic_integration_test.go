//go:build integration

package ifunny

import (
	"context"
	"os"
	"testing"

	"github.com/open-ifunny/ifunny-go/compose"
)

// integrationNickEnv names the environment variable that supplies the
// nickname used to exercise GetUser during integration tests.
const integrationNickEnv = "IFUNNY_INTEGRATION_NICK"

// TestBasicFlow exercises the full unauthenticated path
// (GenerateBasic -> MakeClientBasic -> PrimeBasic -> GetUser by_nick)
// against the live iFunny API for both the Android and iOS device
// profiles. It is gated behind the `integration` build tag and
// includes PrimeBasic's ~15s server-side wait per subtest, so it is
// not part of the default `go test` run.
//
// The target nickname must be provided via the IFUNNY_INTEGRATION_NICK
// environment variable.
//
// Invoke deliberately with:
//
//	IFUNNY_INTEGRATION_NICK=woof go test -tags=integration -run TestBasicFlow -v
func TestBasicFlow(t *testing.T) {
	nick := os.Getenv(integrationNickEnv)
	if nick == "" {
		t.Fatalf("%s must be set to run integration tests", integrationNickEnv)
	}

	cases := []struct {
		name string
		ua   UserAgent
	}{
		{"android", Android{Version: "14"}.UserAgent()},
		{"ios", IOS{Version: "17.5.1"}.UserAgent()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("user-agent: %s", tc.ua.String())

			basic, err := GenerateBasic()
			if err != nil {
				t.Fatalf("GenerateBasic: %v", err)
			}
			if basic == "" {
				t.Fatal("GenerateBasic returned empty token")
			}
			t.Logf("minted basic token (len=%d)", len(basic))

			client, err := MakeClientBasic(basic, tc.ua)
			if err != nil {
				t.Fatalf("MakeClientBasic: %v", err)
			}

			ctx := context.Background()
			t.Log("priming (this takes ~15s)")
			if err := client.PrimeBasic(ctx); err != nil {
				t.Fatalf("PrimeBasic: %v", err)
			}

			user, err := client.GetUser(ctx, compose.UserByNick(nick))
			if err != nil {
				t.Fatalf("GetUser(by_nick %s): %v", nick, err)
			}

			if user.ID == "" {
				t.Fatal("user has empty ID")
			}
			if user.Nick == "" {
				t.Fatal("user has empty Nick")
			}
			t.Logf("fetched user: nick=%s id=%s", user.Nick, user.ID)
		})
	}
}
