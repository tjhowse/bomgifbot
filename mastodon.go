package main

import (
	"context"
	"fmt"
	"time"

	"github.com/tjhowse/go-mastodon"
)

type Mastodon struct {
	c *mastodon.Client
}

func NewMastodon(server, id, secret, access_token string) (*Mastodon, error) {
	m := &Mastodon{}
	m.c = mastodon.NewClient(&mastodon.Config{
		Server:       server,
		ClientID:     id,
		ClientSecret: secret,
		AccessToken:  access_token,
	})
	return m, nil
}

// Posts a status update
func (m *Mastodon) PostStatus(status string) error {
	_, err := m.c.PostStatus(context.Background(), &mastodon.Toot{
		Status: status,
	})
	return err
}

// Gets my last `n` statuses
func (m *Mastodon) GetMyStatuses(n int64) ([]*mastodon.Status, error) {
	if account, err := m.c.GetAccountCurrentUser(context.Background()); err != nil {
		return nil, err
	} else {
		return m.c.GetAccountStatuses(context.Background(), account.ID, &mastodon.Pagination{
			Limit: n,
		})
	}
}

// Posts a status with an image attached
func (m *Mastodon) PostStatusWithImage(status string, filename string) error {
	a, err := m.c.UploadMedia(context.Background(), filename)
	if err != nil {
		return err
	}
	_, err = m.c.PostStatus(context.Background(), &mastodon.Toot{
		Status:   status,
		MediaIDs: []mastodon.ID{a.ID},
	})
	return err
}

// Posts a status with an image attached
func (m *Mastodon) PostStatusWithImageFromBytes(status string, file []byte, visibility string) error {
	fmt.Println("Posting status with image...")
	a, err := m.c.UploadMediaFromBytes(context.Background(), file)
	if err != nil {
		return err
	}
	// Wait until the upload has been processed before posting the status.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) && (m.c.GetMediaStatus(context.Background(), a) != nil) {
		fmt.Println("Waiting for media to be processed...")
		time.Sleep(500 * time.Millisecond)
	}

	_, err = m.c.PostStatus(context.Background(), &mastodon.Toot{
		Status:     status,
		MediaIDs:   []mastodon.ID{a.ID},
		Visibility: visibility,
	})
	return err
}
