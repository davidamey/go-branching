package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"time"

	svg "github.com/ajstarks/svgo"
)

const (
	ArrowSize = 9  // Divisible by 3 for nice-looking arrows
	M         = 5  // Margin, used for a border and for gaps between things
	BarHeight = 20 // Height used for title and week-bar boxes
)

type Diagram struct {
	Title    string    `json:"title"`
	Start    time.Time `json:"start"`
	Weeks    int       `json:"weeks"`
	Width    int       `json:"width"`
	Height   int       `json:"height"`
	Branches Branches  `json:"branches"`
	Merges   []Merge   `json:"merges"`
	Releases []Release `json:"releases"`

	// Computed
	Canvas  *svg.SVG
	YOffset int
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

	diagram.Canvas = svg.New(w)

	return &diagram
}

func (d *Diagram) Draw() {
	d.YOffset = d.Height / (len(d.Branches) + 1)

	wrapperBox := &Box{M, M, d.Width - M, d.Height - M}                                         // Outer box filling diagram with a margin
	titleBox := wrapperBox.NewWithHeight(BarHeight)                                             // Box to hold the title (and potential future sub-title)
	mainBox := wrapperBox.NewAdjusted(100+(wrapperBox.Width()-100)%d.Weeks, titleBox.Y2, 0, -M) // Box holding weekbars and branches etc.
	branchBox := mainBox.NewAdjusted(0, BarHeight+M, 0, -BarHeight-M)                           // Right half of diagram, holds actual branches
	labelBox := &Box{wrapperBox.X1, branchBox.Y1, branchBox.X1, branchBox.Y2}                   // Left half of diagram, holds labels
	releaseBox := &Box{branchBox.X1, branchBox.Y2 - 50, branchBox.X2, branchBox.Y2}             // Box holding deployment labels

	if len(d.Releases) > 0 {
		branchBox.Y2 = releaseBox.Y1 - M
	}

	d.YOffset = branchBox.Height() / len(d.Branches)
	branchBox.Y1 = branchBox.Y1 + d.YOffset/2
	branchBox.Y2 = branchBox.Y2 - d.YOffset/2

	// Draw
	d.Canvas.Start(d.Width, d.Height)
	defer d.Canvas.End()

	d.Canvas.Def()
	for _, t := range [...]BranchType{BranchTypeFeature, BranchTypeMain, BranchTypeRelease, BranchTypeReleaseLive} {
		d.MakeArrowForType(t)
	}
	d.Canvas.DefEnd()

	// d.Canvas.Rect(0, 0, d.Width, d.Height, "stroke:#CCC;fill:none")
	d.DrawTitle(titleBox)
	d.DrawBranches(labelBox, branchBox)
	d.DrawMerges(branchBox)
	d.DrawReleases(releaseBox, branchBox, mainBox)
	d.DrawWeekBars(labelBox, mainBox)

	// d.DrawBoxes(&Box{0, 0, d.Width, d.Height}, wrapperBox, titleBox, mainBox, branchBox, labelBox, releaseBox)
}

func (d *Diagram) DrawBoxes(boxes ...*Box) {
	colours := [...]string{"#000", "#F00", "#0F0", "#00F", "#F0F"}

	for i, b := range boxes {
		d.Canvas.Rect(b.X1, b.Y1, b.Width(), b.Height(),
			`fill="none"`,
			fmt.Sprintf(`stroke="%s"`, colours[i%len(colours)]))
	}
}

func (d *Diagram) DrawTitle(b *Box) {
	d.Canvas.Title(d.Title)
	d.Canvas.Text(b.CenterX(), b.CenterY(), d.Title,
		`font-family="arial"`, `text-anchor="middle"`, `alignment-baseline="central"`)
}

func (d *Diagram) DrawWeekBars(lb, wb *Box) {
	d.Canvas.Rect(lb.X1, wb.Y1, lb.Width(), BarHeight, `fill="#79F"`, `stroke="#000"`)
	d.Canvas.Rect(lb.X1, wb.Y2-BarHeight, lb.Width(), BarHeight, `fill="#79F"`, `stroke="#000"`)

	we := ToWeekEnd(d.Start)
	dw := wb.Width() / d.Weeks
	for i := 0; i < d.Weeks; i++ {
		t := we.AddDate(0, 0, 7*i)
		w := fmt.Sprintf("%d/%d", t.Day(), t.Month())

		FilledRectWithText(d.Canvas, wb.X1+i*dw, wb.Y1, dw, BarHeight, w, "#FFF")           // Top
		FilledRectWithText(d.Canvas, wb.X1+i*dw, wb.Y2-BarHeight, dw, BarHeight, w, "#FFF") // Bottom
	}
}

func (d *Diagram) DrawBranches(lb, bb *Box) {
	for _, b := range d.Branches {
		y := bb.Y1 + b.Order*d.YOffset
		stroke := StrokeForBranch(b)

		d.Canvas.Text(lb.CenterX(), y, b.Name,
			`font-family="arial"`, `text-anchor="middle"`, `alignment-baseline="central"`,
			fmt.Sprintf(`fill="%s"`, b.BranchType.ToColour()))

		x1 := bb.X1 + d.TimeToX(bb, b.Start)
		if x1 > bb.X1 {
			d.Canvas.Line(bb.X1, y, x1, y, `stroke="black"`, `stroke-dasharray="5"`)

			// If we have a parent, draw the branching line
			if b.Parent != nil {
				d.Canvas.Line(x1, bb.Y1+b.Parent.Order*d.YOffset, x1, y, stroke)
			}
		}

		x2 := bb.X2
		if !b.End.IsZero() {
			x2 = bb.X1 + d.TimeToX(bb, b.End)
		}

		// Arrow
		d.Canvas.Line(x1, y, x2-9, y, stroke, ArrowForBranch(b))
	}
}

func (d *Diagram) DrawMerges(bb *Box) {
	for _, m := range d.Merges {
		b1 := d.Branches[m.From]
		b2 := d.Branches[m.To]

		y1 := bb.Y1 + b1.Order*d.YOffset
		y2 := bb.Y1 + b2.Order*d.YOffset
		x := bb.X1 + d.TimeToX(bb, m.Date)

		// Offset y2 for arrow
		if y2 > y1 {
			y2 = y2 - 9
		} else {
			y2 = y2 + 9
		}

		d.Canvas.Line(x, y1, x, y2,
			`stroke-dasharray="2"`,
			StrokeForBranch(b2),
			ArrowForBranch(b2))
	}
}

func (d *Diagram) DrawReleases(rb, bb, mb *Box) {
	w := 80

	for _, r := range d.Releases {
		b := d.Branches[r.From]
		x := rb.X1 + d.TimeToX(rb, r.Date)
		y := bb.Y1 + b.Order*d.YOffset

		d.Canvas.Line(x, mb.Y1, x, mb.Y2, `stroke="#000"`, `stroke-dasharray="2"`)
		d.Canvas.Polygon(
			[]int{x, x + 5, x, x - 5},
			[]int{y + 5, y, y - 5, y},
			`fill="#FF0"`, `stroke="#000"`)
		FilledRectWithText(d.Canvas, x-w/2, rb.Y1, w, rb.Height(), r.Title, "#FF0")
	}
}

func (d *Diagram) TimeToX(b *Box, t time.Time) int {
	x := 0
	if t.After(d.Start) {
		dx := b.Width() / (7 * d.Weeks)
		xOff := int(t.Sub(d.Start).Hours()) / 24
		x = x + xOff*dx
	}

	if x > b.Width() {
		return b.Width()
	} else {
		return x
	}
}

func (d *Diagram) MakeArrowForType(t BranchType) {
	path := fmt.Sprintf("M0,0 L0,%d L%d,%d Z", 2*ArrowSize/3, ArrowSize, ArrowSize/3)
	fill := fmt.Sprintf(`fill="%s"`, t.ToColour())

	d.Canvas.Marker("arrow-"+string(t), 0, 3, 10, 10, `orient="auto"`, `markerUnits="strokeWidth"`)
	d.Canvas.Path(path, fill)
	d.Canvas.MarkerEnd()
}

func RectWithText(c *svg.SVG, x, y, w, h int, t string) {
	FilledRectWithText(c, x, y, w, h, t, "none")
}

func FilledRectWithText(c *svg.SVG, x, y, w, h int, t string, fill string) {
	c.Rect(x, y, w, h, fmt.Sprintf(`fill="%s"`, fill), `stroke="black"`)
	c.Text(x+w/2, y+h/2, t, `font-family="arial"`, `text-anchor="middle"`, `alignment-baseline="central"`)
}

func StrokeForBranch(b *Branch) string {
	return fmt.Sprintf(`stroke="%s"`, b.BranchType.ToColour())
}

func ArrowForBranch(b *Branch) string {
	return fmt.Sprintf(`marker-end="url(#arrow-%s)"`, b.BranchType)
}

func ToWeekEnd(t time.Time) time.Time {
	dayShift := ((time.Friday-t.Weekday())%7 + 7) % 7
	return t.AddDate(0, 0, int(dayShift))
}
