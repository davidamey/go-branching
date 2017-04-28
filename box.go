package main

type Box struct {
	X1, Y1 int // Top left
	X2, Y2 int // Bottom right
}

func (b *Box) Width() int {
	return b.X2 - b.X1
}

func (b *Box) Height() int {
	return b.Y2 - b.Y1
}

func (b *Box) CenterX() int {
	return b.X1 + b.Width()/2
}

func (b *Box) CenterY() int {
	return b.Y1 + b.Height()/2
}

func (b *Box) NewWithWidth(w int) *Box {
	return &Box{b.X1, b.Y1, b.X1 + w, b.Y2}
}

func (b *Box) NewWithHeight(h int) *Box {
	return &Box{b.X1, b.Y1, b.X2, b.Y1 + h}
}

func (b *Box) NewAdjusted(x1, y1, x2, y2 int) *Box {
	return &Box{b.X1 + x1, b.Y1 + y1, b.X2 + x2, b.Y2 + y2}
}
