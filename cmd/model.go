package main

import "time"

type Record struct {
	ID          string    `json:"_id,omitempty" bson:"_id,omitempty"`
	Pub         string    `json:"pub" bson:"pub"`
	Title       string    `json:"title" bson:"title"`
	Page        string    `json:"page" bson:"page"`
	PageCount   int       `json:"page_count" bson:"page_count"`
	Author      string    `json:"author" bson:"author"`
	Creator     string    `json:"creator" bson:"creator"`
	AuthorName  string    `json:"author_name" bson:"author_name"`
	Ordinal     string    `json:"ordinal" bson:"ordinal"`
	Stream      string    `json:"stream" bson:"stream"`
	StreamName  string    `json:"stream_name" bson:"stream_name"`
	Affiliation string    `json:"affiliation" bson:"affiliation"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
	Source      string    `json:"source" bson:"source"`
}