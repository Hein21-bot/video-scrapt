package models

import (
	"math/rand"
)

type ShareDoc struct {
	Code      string `bson:"code"      json:"code"`
	VideoID   string `bson:"video_id"  json:"video_id"`
	PageURL   string `bson:"page_url"  json:"-"`
	Title     string `bson:"title"     json:"title"`
	Thumbnail string `bson:"thumbnail" json:"thumbnail"`
	Site      string `bson:"site"      json:"site"`
}

const codeChars = "abcdefghijklmnopqrstuvwxyz0123456789"

func RandCode(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = codeChars[rand.Intn(len(codeChars))]
	}
	return string(b)
}
