package main

import (
	"fmt"
	"io"

	"time"

	svg "github.com/ajstarks/svgo"
)

type Diagram struct {
	Canvas     *svg.SVG
	Width      int
	Height     int
	LabelWidth int
	Start      time.Time
	Weeks      int
	Branches   []*Branch
}

func NewDiagram(w io.Writer, width, height int, branches []*Branch) *Diagram {
	canvas := svg.New(w)
	canvas.Start(width, height)
	canvas.Rect(0, 0, width, height, "stroke:#CCC;fill:#FFF")

	weeks := 15
	labelWidth := 100 + (width-100)%weeks

	return &Diagram{
		canvas,
		width,
		height,
		labelWidth,
		time.Now(),
		weeks,
		branches,
	}
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

func (d *Diagram) GetBranchYOffset() int {
	return d.Height / (len(d.Branches) + 1)
}

func (d *Diagram) DrawWeekBars() {
	d.Canvas.Rect(0, 0, d.LabelWidth, 20, "fill:#79F;stroke:black")
	d.Canvas.Text(0+d.LabelWidth/2, 15, "Past", "font-family:arial;text-anchor:middle")
	d.Canvas.Rect(0, d.Height-20, d.LabelWidth, 20, "fill:#79F;stroke:black")
	d.Canvas.Text(0+d.LabelWidth/2, d.Height-5, "Past", "font-family:arial;text-anchor:middle")

	we := ToWeekEnd(d.Start)
	dw := (d.Width - d.LabelWidth) / d.Weeks
	for i := 0; i < d.Weeks; i++ {
		t := we.AddDate(0, 0, 7*i)
		w := fmt.Sprintf("%d/%d", t.Day(), t.Month())

		// Top
		d.Canvas.Rect(d.LabelWidth+i*dw, 0, dw, 20, "fill:none;stroke:black;stroke-width:1")
		d.Canvas.Text(d.LabelWidth+i*dw+dw/2, 15, w, "font-family:arial;text-anchor:middle")

		// Bottom
		d.Canvas.Rect(d.LabelWidth+i*dw, d.Height-20, dw, 20, "fill:none;stroke:black;stroke-width:1")
		d.Canvas.Text(d.LabelWidth+i*dw+dw/2, d.Height-5, w, "font-family:arial;text-anchor:middle")
	}
}

func (d *Diagram) DrawBranch(b *Branch) {
	y := b.Order * d.GetBranchYOffset()

	d.Canvas.Text(10, y+5, b.Name, "font-family:arial;")

	var stroke string
	switch b.BranchType {
	case BranchTypeFeature:
		stroke = "stroke:red;"
	case BranchTypeRelease:
		stroke = "stroke:green;"
	case BranchTypeMain:
		stroke = "stroke:black;stroke-width:1;"
	default:
		stroke = "stroke:black;"
	}

	dashStroke := "stroke-dasharray:5,5;"

	x := d.LabelWidth
	if b.Created.After(d.Start) {
		dx := (d.Width - d.LabelWidth) / (7 * d.Weeks)
		xOff := int(b.Created.Sub(d.Start).Hours()) / 24
		x = x + xOff*dx
		d.Canvas.Line(d.LabelWidth, y, x, y, stroke+dashStroke)

		// If we have a parent, draw the branching line
		if b.Parent != nil {
			d.Canvas.Line(x, b.Parent.Order*d.GetBranchYOffset(), x, y, stroke)
		}
	}

	// Arrow trunk
	d.Canvas.Line(x, y, d.Width, y, stroke)

	// Arrow tip
	d.Canvas.Line(d.Width, y, d.Width-15, y-10, stroke)
	d.Canvas.Line(d.Width, y, d.Width-15, y+10, stroke)
}

func (d *Diagram) End() {
	d.Canvas.End()
}
