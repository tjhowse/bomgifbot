package main

import (
	"context"
	"os"

	"github.com/mattn/go-mastodon"
)

type Mastodon struct {
	c *mastodon.Client
}

func NewMastodon(server, id, secret string) (*Mastodon, error) {
	m := &Mastodon{}
	m.c = mastodon.NewClient(&mastodon.Config{
		Server:       server,
		ClientID:     id,
		ClientSecret: secret,
	})
	err := m.c.Authenticate(context.Background(), os.Getenv("MASTODON_USER_EMAIL"), os.Getenv("MASTODON_USER_PASSWORD"))
	if err != nil {
		return nil, err
	}
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

// Uploads an image to the server and returns an attachment pointer
func (m *Mastodon) UploadImage(filename string) (*mastodon.Attachment, error) {
	return m.c.UploadMedia(context.Background(), filename)
}

// Posts a status with an image attached
func (m *Mastodon) PostStatusWithImage(status string, filename string) error {
	a, err := m.UploadImage(filename)
	if err != nil {
		return err
	}
	_, err = m.c.PostStatus(context.Background(), &mastodon.Toot{
		Status:   status,
		MediaIDs: []mastodon.ID{a.ID},
	})
	return err
}
