package main

import (
	"time"
)

type Release struct {
	Date        time.Time `json:"date"`
	From        string    `json:"from"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
}
