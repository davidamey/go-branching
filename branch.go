package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type BranchType uint8

const (
	BranchTypeMain BranchType = iota
	BranchTypeFeature
	BranchTypeRelease
)

type Branch struct {
	Order      int        `json:"-"`
	Name       string     `json:"name"`
	ParentName string     `json:"parent"`
	Parent     *Branch    `json:"-"`
	Created    time.Time  `json:"created"`
	BranchType BranchType `json:"type"`
}

func LoadBranches() []*Branch {
	var branches []*Branch

	r, _ := os.Open("branches.json")
	defer r.Close()

	d := json.NewDecoder(r)
	err := d.Decode(&branches)
	if err != nil {
		fmt.Println(err)
	}

	// Create a name -> *Branch map to save repeated looping
	hash := make(map[string]*Branch)
	for i, b := range branches {
		hash[b.Name] = branches[i]
	}

	for i, b := range branches {
		branches[i].Order = i + 1
		branches[i].Parent = hash[b.ParentName]
	}

	return branches

	// dev := &Branch{2, "Dev", "", nil, time.Now(), BranchTypeMain}
	// f1 := &Branch{1, "Feature 1", "Dev", dev, time.Now().AddDate(0, 0, 14), BranchTypeFeature}
	// r11 := &Branch{3, "Release 11", "Dev", dev, time.Now(), BranchTypeRelease}

	// return []*Branch{f1, dev, r11}
}
