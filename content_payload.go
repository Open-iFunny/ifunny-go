package ifunny

// Payload is implemented by every kind-specific payload struct embedded in
// Content (Content.Pic, Content.VideoClip, Content.Gif, ...). On its own it
// says nothing about what the payload contains — it exists so Content.Payload
// can hand back whatever is populated as a single value, and so the method
// set is sealed to the payload types declared in this package.
//
// The useful information comes from type-asserting a Payload against the
// narrower capability interfaces below. Not every kind implements every
// capability: what a "pic" and a "video_clip" have in common isn't a shared
// set of fields, it's that both answer a subset of the same questions
// ("where's the preview image?", "where's the playable media?", "how big is
// it?"). Modeling that as a handful of single-method interfaces (in the
// spirit of io.Reader/io.Writer) lets callers ask only the question they
// care about instead of switching over every ContentKind.
type Payload interface {
	// sealed prevents types outside this package from satisfying Payload.
	sealed()
}

// Media is implemented by payloads whose data is a single URL pointing at
// the actual media that was posted: images (Pic, Comics, Meme), the
// original source of an imported video (Video), gifs (Gif — covers both
// CONTENT_GIF and CONTENT_GIF_CAPTION), and third-party embeds (App).
// VideoClip, Vine, and Coub don't implement Media: the API only gives us a
// still preview for them (see Preview), not a direct media URL.
type Media interface {
	Payload
	URL() string
}

// Preview is implemented by payloads that carry a still-frame preview image
// distinct from their own media: VideoClip, Vine, Coub, and Gif.
type Preview interface {
	Payload
	PreviewURL() string
}

// Sized is implemented by payloads that report their own encoded size in
// bytes: VideoClip, Vine, Coub, and Gif. Kinds without a byte count in the
// API response (Pic-family, Caption, App) don't implement it.
type Sized interface {
	Payload
	Size() int
}

// Captioned is implemented by payloads that carry caption text composited
// over other content: Caption, and Gif (meaningful only when the owning
// Content's Type is CONTENT_GIF_CAPTION; empty for plain gifs).
type Captioned interface {
	Payload
	Caption() string
}

// PayloadPic is the payload for CONTENT_PIC, CONTENT_COMICS, and
// CONTENT_MEME: a single webp image. The three kinds differ only in which
// in-app tool produced the image, not in shape.
type PayloadPic struct {
	WebpURL string `json:"webp_url"`
}

func (*PayloadPic) sealed()       {}
func (p *PayloadPic) URL() string { return p.WebpURL }

// PayloadVideoClip is the payload for CONTENT_VIDEO_CLIP, the most common
// video kind.
type PayloadVideoClip struct {
	ScreenURL  string `json:"screen_url"` // JPEG preview frame
	SourceType string `json:"source_type"`
	LogoURL    string `json:"logo_url,omitempty"` // source logo, present when SourceType != "user"
	Bytes      int    `json:"bytes"`
	Duration   int    `json:"duration"` // seconds
}

func (*PayloadVideoClip) sealed()              {}
func (p *PayloadVideoClip) PreviewURL() string { return p.ScreenURL }
func (p *PayloadVideoClip) Size() int          { return p.Bytes }

// PayloadVideo is the payload for CONTENT_VIDEO: a video imported from an
// external URL (e.g. YouTube), as opposed to one recorded/uploaded directly
// (CONTENT_VIDEO_CLIP).
type PayloadVideo struct {
	SourceURL string `json:"url"`
	Duration  int    `json:"duration"` // seconds
	Length    int    `json:"length"`   // seconds; observed equal to Duration
}

func (*PayloadVideo) sealed()       {}
func (p *PayloadVideo) URL() string { return p.SourceURL }

// PayloadVine is the payload for CONTENT_VINE: a legacy import from the
// defunct Vine app.
type PayloadVine struct {
	ScreenURL string `json:"screen_url"`
	Bytes     int    `json:"bytes"`
}

func (*PayloadVine) sealed()              {}
func (p *PayloadVine) PreviewURL() string { return p.ScreenURL }
func (p *PayloadVine) Size() int          { return p.Bytes }

// PayloadCoub is the payload for CONTENT_COUB: a video meme imported from
// coub.com.
type PayloadCoub struct {
	ScreenURL    string `json:"screen_url"`
	Bytes        int    `json:"bytes"`
	TrackbackURL string `json:"trackback_url"` // original Coub source
	Duration     int    `json:"duration"`      // seconds
}

func (*PayloadCoub) sealed()              {}
func (p *PayloadCoub) PreviewURL() string { return p.ScreenURL }
func (p *PayloadCoub) Size() int          { return p.Bytes }

// PayloadGif is the payload for both CONTENT_GIF and CONTENT_GIF_CAPTION;
// the API uses the same "gif" JSON key for either kind. CaptionText is only
// populated for CONTENT_GIF_CAPTION.
type PayloadGif struct {
	ScreenURL   string `json:"screen_url"` // JPEG preview frame
	Bytes       int    `json:"bytes"`
	Mp4URL      string `json:"mp4_url"`
	Mp4Bytes    int    `json:"mp4_bytes"`
	WebmURL     string `json:"webm_url,omitempty"`
	WebmBytes   int    `json:"webm_bytes"`
	CaptionText string `json:"caption_text,omitempty"` // only set for CONTENT_GIF_CAPTION
}

func (*PayloadGif) sealed()              {}
func (p *PayloadGif) URL() string        { return p.Mp4URL }
func (p *PayloadGif) PreviewURL() string { return p.ScreenURL }
func (p *PayloadGif) Size() int          { return p.Bytes }
func (p *PayloadGif) Caption() string    { return p.CaptionText }

// PayloadCaption is the payload for CONTENT_CAPTION: a custom caption
// composited onto an image meme. Unlike Pic, it carries no media URL of its
// own — the underlying image lives on the Content itself.
type PayloadCaption struct {
	CaptionText string `json:"caption_text"`
}

func (*PayloadCaption) sealed()           {}
func (p *PayloadCaption) Caption() string { return p.CaptionText }

// PayloadApp is the payload for CONTENT_APP: an interactive/iframe card.
// Deprecated by iFunny; kept for decoding historical content.
type PayloadApp struct {
	SourceURL       string `json:"url"`
	IsScrollAllowed bool   `json:"is_scroll_allowed"`
}

func (*PayloadApp) sealed()       {}
func (p *PayloadApp) URL() string { return p.SourceURL }
