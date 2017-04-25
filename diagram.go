package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"time"

	svg "github.com/ajstarks/svgo"
)

type Diagram struct {
	Title    string    `json:"title"`
	Start    time.Time `json:"start"`
	Weeks    int       `json:"weeks"`
	Width    int       `json:"width"`
	Height   int       `json:"height"`
	Branches Branches  `json:"branches"`

	// Computed
	Canvas     *svg.SVG
	LabelWidth int
	YOffset    int
}

func NewDiagram(cfg string, w io.Writer) *Diagram {
	var diagram Diagram

	r, _ := os.Open(cfg)
	defer r.Close()

	d := json.NewDecoder(r)
	err := d.Decode(&diagram)
	if err != nil {
		fmt.Println(err)
	}

	diagram.LabelWidth = 100 + (diagram.Width-100)%diagram.Weeks
	diagram.YOffset = diagram.Height / (len(diagram.Branches) + 1)

	diagram.Canvas = svg.New(w)

	return &diagram
}

func ToWeekEnd(t time.Time) time.Time {
	switch t.Weekday() {
	case time.Sunday:
		return t.AddDate(0, 0, 5)
	case time.Monday:
		return t.AddDate(0, 0, 4)
	case time.Tuesday:
		return t.AddDate(0, 0, 3)
	case time.Wednesday:
		return t.AddDate(0, 0, 2)
	case time.Thursday:
		return t.AddDate(0, 0, 1)
	case time.Saturday:
		return t.AddDate(0, 0, 6)
	default:
		return t
	}
}

func (d *Diagram) Draw() {
	d.Canvas.Start(d.Width, d.Height)
	defer d.Canvas.End()

	d.Canvas.Def()
	d.Canvas.Marker("arrow-feature", 0, 3, 10, 10, `orient="auto"`, `markerUnits="strokeWidth"`)
	d.Canvas.Path("M0,0 L0,6 L9,3 z", `fill="red"`)
	d.Canvas.MarkerEnd()
	d.Canvas.Marker("arrow-main", 0, 3, 10, 10, `orient="auto"`, `markerUnits="strokeWidth"`)
	d.Canvas.Path("M0,0 L0,6 L9,3 z", `fill="#000"`)
	d.Canvas.MarkerEnd()
	d.Canvas.Marker("arrow-release", 0, 3, 10, 10, `orient="auto"`, `markerUnits="strokeWidth"`)
	d.Canvas.Path("M0,0 L0,6 L9,3 z", `fill="green"`)
	d.Canvas.MarkerEnd()
	d.Canvas.DefEnd()

	d.Canvas.Rect(0, 0, d.Width, d.Height, "stroke:#CCC;fill:#FFF")
	d.DrawWeekBars()

	for _, b := range d.Branches {
		d.DrawBranch(b)
	}
}

func (d *Diagram) DrawWeekBars() {
	boxStyle := []string{`fill="none"`, `stroke="black"`}
	textStyle := []string{`font-family="arial"`, `text-anchor="middle"`}

	d.Canvas.Rect(0, 0, d.LabelWidth, 20, `fill="#79F"`, `stroke="black"`)
	d.Canvas.Text(0+d.LabelWidth/2, 15, "Past", textStyle...)
	d.Canvas.Rect(0, d.Height-20, d.LabelWidth, 20, `fill="#79F"`, `stroke="black"`)
	d.Canvas.Text(0+d.LabelWidth/2, d.Height-5, "Past", textStyle...)

	we := ToWeekEnd(d.Start)
	dw := (d.Width - d.LabelWidth) / d.Weeks
	for i := 0; i < d.Weeks; i++ {
		t := we.AddDate(0, 0, 7*i)
		w := fmt.Sprintf("%d/%d", t.Day(), t.Month())

		// Top
		d.Canvas.Rect(d.LabelWidth+i*dw, 0, dw, 20, boxStyle...)
		d.Canvas.Text(d.LabelWidth+i*dw+dw/2, 15, w, textStyle...)

		// Bottom
		d.Canvas.Rect(d.LabelWidth+i*dw, d.Height-20, dw, 20, boxStyle...)
		d.Canvas.Text(d.LabelWidth+i*dw+dw/2, d.Height-5, w, textStyle...)
	}
}

func (d *Diagram) DrawBranch(b *Branch) {
	y := b.Order * d.YOffset

	d.Canvas.Text(10, y+5, b.Name, "font-family:arial;")

	stroke := fmt.Sprintf(`stroke="%s"`, b.BranchType.ToColour())
	dashStroke := `stroke-dasharray="5,5"`

	x := d.LabelWidth
	if b.Created.After(d.Start) {
		dx := (d.Width - d.LabelWidth) / (7 * d.Weeks)
		xOff := int(b.Created.Sub(d.Start).Hours()) / 24
		x = x + xOff*dx
		d.Canvas.Line(d.LabelWidth, y, x, y, stroke, dashStroke)

		// If we have a parent, draw the branching line
		if b.Parent != nil {
			d.Canvas.Line(x, b.Parent.Order*d.YOffset, x, y, stroke)
		}
	}

	// Arrow trunk
	d.Canvas.Line(x, y, d.Width-10, y, stroke, fmt.Sprintf(`marker-end="url(#arrow-%s)"`, b.BranchType))
}
