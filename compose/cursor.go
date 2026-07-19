package compose

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

// IDs is a page token represented as its underlying content-ID exclusion set.
// The collective feed's cursor is not opaque: it is base64std(zlib("id1,id2,…")),
// a comma-separated list of every content ID already shown this session, which
// the server treats as an exclusion set. IDs lets us mint and reshape that token
// directly (see [TailPager]). Its String is the encoded wire form.
type IDs []string

// String encodes the ID list as the server's cursor form: the comma-joined IDs,
// zlib-compressed, then standard-base64'd. Re-encoding a decoded cursor is
// decode-equal but not byte-equal to the server's — the server re-parses the
// CSV, so only the ID list matters.
func (ids IDs) String() string {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	// zlib.Writer.Write to an in-memory buffer does not fail.
	w.Write([]byte(strings.Join(ids, ",")))
	w.Close()
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

// DecodeIDs is the inverse of [IDs.String]: it base64-decodes then zlib-inflates
// a cursor into its content-ID list. It errors on a malformed token (bad base64
// or non-zlib payload).
func DecodeIDs(token string) (IDs, error) {
	raw, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}

	r, err := zlib.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if len(out) == 0 {
		return IDs{}, nil
	}
	return IDs(strings.Split(string(out), ",")), nil
}

// TailPager returns a pager transform that keeps only the last n IDs of a
// cursor, preserving order. This dodges the collective size cliff: the server
// grows its exclusion set unbounded and 400s on its own oversized cursor, but a
// short trailing window stays well under the limit while still excluding
// recently-seen content. A malformed cursor is surfaced as an error rather than
// silently forwarded. n <= 0 disables trimming and passes the token through.
func TailPager(n int) func(string) (string, error) {
	if n <= 0 {
		return func(token string) (string, error) { return token, nil }
	}

	return func(token string) (string, error) {
		ids, err := DecodeIDs(token)
		if err != nil {
			return "", fmt.Errorf("tail pager decoding cursor: %w", err)
		}

		if len(ids) > n {
			ids = ids[len(ids)-n:]
		}
		return ids.String(), nil
	}
}
