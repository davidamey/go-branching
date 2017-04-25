package main

import (
	"time"
)

type Merge struct {
	Date time.Time `json:"date"`
	From string    `json:"from"`
	To   string    `json:"to"`
}
