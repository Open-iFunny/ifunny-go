package ifunny

import (
	"github.com/open-ifunny/ifunny-go/compose"
)

type Comment struct {
	ID           string `json:"id"`
	CID          string `json:"cid"`
	State        string `json:"state"`
	Date         int    `json:"date"`
	Text         string `json:"text"`
	IsReply      bool   `json:"is_reply"`
	IsSmiled     bool   `json:"is_smiled"`
	IsUnsmiled   bool   `json:"is_unsmiled"`
	IsEdited     bool   `json:"is_edited"`
	RootCommID   string `json:"root_comm_id"`
	ParentCommID string `json:"parent_comm_id"`
	Depth        int    `json:"depth"`
	User         User   `json:"user"`
	Num          struct {
		Smiles   int `json:"smiles"`
		Unsmiles int `json:"unsmiles"`
		Replies  int `json:"replies"`
	} `json:"num"`
	Attachments struct {
		MentionUser []struct {
			ID           string `json:"id"`
			UserID       string `json:"user_id"`
			Nick         string `json:"nick"`
			Start        int    `json:"start_index"`
			Stop         int    `json:"stop_index"`
			OriginalNick string `json:"original_nick"`
		} `json:"mention_user"`
	} `json:"attachments"`
	ContentThumbs struct {
		URL                 string `json:"url"`
		LargeURL            string `json:"large_url"`
		X640URL             string `json:"x640_url"`
		WebpURL             string `json:"webp_url"`
		LargeWebpURL        string `json:"large_webp_url"`
		X640WebpURL         string `json:"x640_webp_url"`
		ProportionalURL     string `json:"proportional_url"`
		ProportionalWebpURL string `json:"proportional_webp_url"`
		ProportionalSize    struct {
			Width  int `json:"w"`
			Height int `json:"h"`
		} `json:"proportional_size"`
	} `json:"content_thumbs"`
}

func (client *Client) GetCommentPage(request compose.Request) (*Page[Comment], error) {
	content := new(struct {
		Data struct {
			Comments Page[Comment] `json:"comments"`
		}
	})

	err := client.RequestJSON(request, content)
	return &content.Data.Comments, err
}

func (client *Client) IterComments(id string) <-chan Result[*Comment] {
	return iterFrom(client, func(limit int, page compose.Page[string]) compose.Request { return compose.Comments(id, limit, page) }, client.GetCommentPage)
}
