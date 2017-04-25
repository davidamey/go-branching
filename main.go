package main

import (
	"log"
	"os"
)

func main() {
	f, err := os.Create("out.svg")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	d := NewDiagram(f, 891, 630, LoadBranches())

	for _, b := range d.Branches {
		d.DrawBranch(b)
	}

	d.DrawWeekBars()
	d.Canvas.End()
}
