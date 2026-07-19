package compose

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"reflect"
	"testing"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	ids := IDs{"aBc123", "dEf456", "gHi789"}

	got, err := DecodeIDs(ids.String())
	if err != nil {
		t.Fatalf("DecodeIDs: %v", err)
	}
	if !reflect.DeepEqual(got, ids) {
		t.Fatalf("round trip: got %v, want %v", got, ids)
	}
}

func TestEncodeIsZlibBase64(t *testing.T) {
	// The server's cursor is base64std(zlib(csv)); a default-level zlib stream
	// starts with the header bytes 0x78 0x9c. This pins our encoding to that form.
	token := (IDs{"one", "two"}).String()

	raw, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		t.Fatalf("token is not valid base64: %v", err)
	}
	if len(raw) < 2 || raw[0] != 0x78 || raw[1] != 0x9c {
		t.Fatalf("payload is not a zlib stream: % x", raw)
	}
}

func TestDecodeCanonicalZlib(t *testing.T) {
	// Decode a token produced by the standard library's zlib writer (i.e. what the
	// server emits), independent of our own String(), to prove interoperability.
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	w.Write([]byte("id1,id2,id3"))
	w.Close()
	token := base64.StdEncoding.EncodeToString(buf.Bytes())

	got, err := DecodeIDs(token)
	if err != nil {
		t.Fatalf("DecodeIDs: %v", err)
	}
	want := IDs{"id1", "id2", "id3"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestDecodeErrors(t *testing.T) {
	if _, err := DecodeIDs("not valid base64!!!"); err == nil {
		t.Fatal("expected error on bad base64")
	}

	// Valid base64 but not a zlib stream.
	if _, err := DecodeIDs(base64.StdEncoding.EncodeToString([]byte("hello"))); err == nil {
		t.Fatal("expected error on non-zlib payload")
	}
}

func TestTailPagerKeepsLastN(t *testing.T) {
	full := IDs{"a", "b", "c", "d", "e"}
	token, err := TailPager(2)(full.String())
	if err != nil {
		t.Fatalf("TailPager: %v", err)
	}

	got, err := DecodeIDs(token)
	if err != nil {
		t.Fatalf("DecodeIDs: %v", err)
	}
	want := IDs{"d", "e"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("tail: got %v, want %v", got, want)
	}
}

func TestTailPagerShorterThanN(t *testing.T) {
	full := IDs{"a", "b"}
	token, err := TailPager(10)(full.String())
	if err != nil {
		t.Fatalf("TailPager: %v", err)
	}

	got, err := DecodeIDs(token)
	if err != nil {
		t.Fatalf("DecodeIDs: %v", err)
	}
	if !reflect.DeepEqual(got, full) {
		t.Fatalf("tail shorter than n: got %v, want %v", got, full)
	}
}

func TestTailPagerZeroIsIdentity(t *testing.T) {
	// n <= 0 disables trimming and passes the token through without decoding,
	// so it never errors even on an opaque value.
	got, err := TailPager(0)("anything")
	if err != nil {
		t.Fatalf("n=0 should not error: %v", err)
	}
	if got != "anything" {
		t.Fatalf("n=0 should be identity, got %q", got)
	}
}

func TestTailPagerErrorsOnMalformed(t *testing.T) {
	// A trimming pager cannot decode a malformed cursor, so it surfaces the
	// error rather than forwarding a bad token.
	if _, err := TailPager(5)("not-a-cursor!!!"); err == nil {
		t.Fatal("expected an error on a malformed cursor")
	}
}
