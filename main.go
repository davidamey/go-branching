package main

import (
	"log"
	"os"
)

func main() {
	w, err := os.Create("out.svg")
	if err != nil {
		log.Fatal(err)
	}
	defer w.Close()

	d := NewDiagram("ngba.json", w)
	d.Draw()
}
