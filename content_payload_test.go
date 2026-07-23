package ifunny

import (
	"encoding/json"
	"testing"
)

// TestContent_Payload_DecodesEveryKnownKind decodes a minimal JSON payload
// for every ContentKind with a documented shape and confirms Content.Payload
// returns the matching populated struct.
func TestContent_Payload_DecodesEveryKnownKind(t *testing.T) {
	cases := []struct {
		name string
		json string
		want Payload
	}{
		{"pic", `{"type":"pic","pic":{"webp_url":"https://x/pic.webp"}}`,
			&PayloadPic{WebpURL: "https://x/pic.webp"}},
		{"comics", `{"type":"comics","comics":{"webp_url":"https://x/comic.webp"}}`,
			&PayloadPic{WebpURL: "https://x/comic.webp"}},
		{"mem", `{"type":"mem","mem":{"webp_url":"https://x/meme.webp"}}`,
			&PayloadPic{WebpURL: "https://x/meme.webp"}},
		{"video_clip", `{"type":"video_clip","video_clip":{"screen_url":"https://x/s.jpg","bytes":10,"source_type":"user","duration":5}}`,
			&PayloadVideoClip{ScreenURL: "https://x/s.jpg", Bytes: 10, SourceType: "user", Duration: 5}},
		{"video", `{"type":"video","video":{"url":"https://x/v.mp4","duration":5,"length":5}}`,
			&PayloadVideo{SourceURL: "https://x/v.mp4", Duration: 5, Length: 5}},
		{"vine", `{"type":"vine","vine":{"screen_url":"https://x/s.jpg","bytes":10}}`,
			&PayloadVine{ScreenURL: "https://x/s.jpg", Bytes: 10}},
		{"coub", `{"type":"coub","coub":{"screen_url":"https://x/s.jpg","bytes":10,"trackback_url":"https://coub.com/x","duration":5}}`,
			&PayloadCoub{ScreenURL: "https://x/s.jpg", Bytes: 10, TrackbackURL: "https://coub.com/x", Duration: 5}},
		{"gif", `{"type":"gif","gif":{"screen_url":"https://x/s.jpg","bytes":10,"mp4_url":"https://x/g.mp4","mp4_bytes":20}}`,
			&PayloadGif{ScreenURL: "https://x/s.jpg", Bytes: 10, Mp4URL: "https://x/g.mp4", Mp4Bytes: 20}},
		{"gif_caption", `{"type":"gif_caption","gif":{"screen_url":"https://x/s.jpg","bytes":10,"mp4_url":"https://x/g.mp4","mp4_bytes":20,"caption_text":"lol"}}`,
			&PayloadGif{ScreenURL: "https://x/s.jpg", Bytes: 10, Mp4URL: "https://x/g.mp4", Mp4Bytes: 20, CaptionText: "lol"}},
		{"caption", `{"type":"caption","caption":{"caption_text":"lol"}}`,
			&PayloadCaption{CaptionText: "lol"}},
		{"app", `{"type":"app","app":{"url":"https://x/app","is_scroll_allowed":true}}`,
			&PayloadApp{SourceURL: "https://x/app", IsScrollAllowed: true}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var c Content
			if err := json.Unmarshal([]byte(tc.json), &c); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}

			got := c.Payload()
			if got == nil {
				t.Fatal("Payload() = nil, want a populated payload")
			}
			if gotJSON, wantJSON := mustJSON(t, got), mustJSON(t, tc.want); gotJSON != wantJSON {
				t.Errorf("Payload() = %s, want %s", gotJSON, wantJSON)
			}
		})
	}
}

// TestContent_Payload_UndocumentedKinds confirms that CONTENT_OLD,
// CONTENT_DEM, and CONTENT_SPECIAL — which have no documented payload shape
// — decode without error and yield a nil Payload rather than panicking or
// fabricating data.
func TestContent_Payload_UndocumentedKinds(t *testing.T) {
	for _, kind := range []ContentKind{CONTENT_OLD, CONTENT_DEM, CONTENT_SPECIAL} {
		var c Content
		body := `{"type":"` + string(kind) + `"}`
		if err := json.Unmarshal([]byte(body), &c); err != nil {
			t.Fatalf("Unmarshal(%s): %v", kind, err)
		}
		if p := c.Payload(); p != nil {
			t.Errorf("Payload() for %s = %#v, want nil", kind, p)
		}
	}
}

// TestPayload_CapabilityInterfaces spot-checks which capability interfaces
// (Media, Preview, Sized, Captioned) each payload kind is expected to
// implement, since that's the whole point of splitting them out instead of
// one bloated interface.
func TestPayload_CapabilityInterfaces(t *testing.T) {
	pic := &PayloadPic{WebpURL: "https://x/p.webp"}
	assertImplements[Media](t, "PayloadPic", pic)

	clip := &PayloadVideoClip{ScreenURL: "https://x/s.jpg"}
	assertImplements[Preview](t, "PayloadVideoClip", clip)
	assertImplements[Sized](t, "PayloadVideoClip", clip)

	video := &PayloadVideo{SourceURL: "https://x/v.mp4"}
	assertImplements[Media](t, "PayloadVideo", video)

	gif := &PayloadGif{Mp4URL: "https://x/g.mp4", ScreenURL: "https://x/s.jpg", CaptionText: "lol"}
	assertImplements[Media](t, "PayloadGif", gif)
	assertImplements[Preview](t, "PayloadGif", gif)
	assertImplements[Sized](t, "PayloadGif", gif)
	assertImplements[Captioned](t, "PayloadGif", gif)

	caption := &PayloadCaption{CaptionText: "lol"}
	assertImplements[Captioned](t, "PayloadCaption", caption)

	app := &PayloadApp{SourceURL: "https://x/app"}
	assertImplements[Media](t, "PayloadApp", app)
}

func assertImplements[I any](t *testing.T, name string, p Payload) {
	t.Helper()
	if _, ok := p.(I); !ok {
		var zero I
		t.Errorf("%s does not implement %T", name, zero)
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	return string(b)
}
