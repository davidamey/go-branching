package main

import (
	"encoding/json"
	"time"
)

type BranchType string

const (
	BranchTypeMain        BranchType = "main"
	BranchTypeFeature                = "feature"
	BranchTypeRelease                = "release"
	BranchTypeReleaseLive            = "live"
)

type Branch struct {
	Order      int        `json:"-"`
	Name       string     `json:"name"`
	ParentName string     `json:"parent"`
	Parent     *Branch    `json:"-"`
	Start      time.Time  `json:"start"`
	End        time.Time  `json:"end"`
	BranchType BranchType `json:"type"`
}

type Branches map[string]*Branch

func (bt *BranchType) ToColour() string {
	switch *bt {
	case BranchTypeFeature:
		return "red"
	case BranchTypeRelease:
		return "green"
	case BranchTypeReleaseLive:
		return "purple"
	case BranchTypeMain:
		return "#000"
	default:
		return "#000"
	}
}

func (this *Branches) UnmarshalJSON(b []byte) error {
	var branches []*Branch

	err := json.Unmarshal(b, &branches)
	if err != nil {
		return err
	}

	// Create a name -> *Branch map to save repeated looping
	*this = make(map[string]*Branch)
	for i, b := range branches {
		(*this)[b.Name] = branches[i]
	}

	for i, b := range branches {
		branches[i].Order = i + 1
		branches[i].Parent = (*this)[b.ParentName]
	}

	return nil
}

// func (branch *Branch) UnmarshalJSON(b []byte) error {
// 	fmt.Println("Branch Unmarshal")
// 	return nil
// }
