// Package layout provides the layout engine for Gocko
// Flexbox implementation follows CSS Flexible Box Layout Module Level 1
// https://www.w3.org/TR/css-flexbox-1/
package layout

import (
	"go-browser/gocko/css/values"
)

// =============================================================================
// FLEXBOX LAYOUT ENGINE
// Implements CSS Flexible Box Layout Module Level 1
// =============================================================================

// FlexItem represents a child item in a flex container
type FlexItem struct {
	// Node reference
	Node interface{}

	// Base sizes
	MainSize     float64 // Size along main axis
	CrossSize    float64 // Size along cross axis
	FlexBaseSize float64 // Computed flex-basis

	// Flex values
	FlexGrow   float64
	FlexShrink float64

	// Computed positions (output)
	MainPos  float64
	CrossPos float64

	// Computed final sizes (output)
	FinalMainSize  float64
	FinalCrossSize float64

	// Margins
	MarginMainStart  float64
	MarginMainEnd    float64
	MarginCrossStart float64
	MarginCrossEnd   float64

	// Alignment
	AlignSelf string
}

// FlexContainer represents a flex container
type FlexContainer struct {
	// Container dimensions
	Width  float64
	Height float64

	// Flex container properties
	Direction      string // row, column, row-reverse, column-reverse
	Wrap           string // nowrap, wrap, wrap-reverse
	JustifyContent string // flex-start, flex-end, center, space-between, space-around, space-evenly
	AlignItems     string // flex-start, flex-end, center, stretch, baseline
	AlignContent   string // flex-start, flex-end, center, stretch, space-between, space-around
	Gap            float64

	// Items
	Items []*FlexItem

	// Computed flex lines (for wrap)
	Lines []FlexLine
}

// FlexLine represents a line of flex items (for wrap)
type FlexLine struct {
	Items      []*FlexItem
	MainSize   float64 // Total main size of items
	CrossSize  float64 // Max cross size of items
	MainStart  float64 // Position where this line starts on main axis
	CrossStart float64 // Position where this line starts on cross axis
}

// NewFlexContainer creates a new flex container
func NewFlexContainer(width, height float64, style *values.ComputedStyle) *FlexContainer {
	return &FlexContainer{
		Width:          width,
		Height:         height,
		Direction:      style.FlexDirection,
		Wrap:           style.FlexWrap,
		JustifyContent: style.JustifyContent,
		AlignItems:     style.AlignItems,
		AlignContent:   style.AlignContent,
		Gap:            style.Gap.Resolve(values.DefaultContext()),
	}
}

// AddItem adds a flex item to the container
func (fc *FlexContainer) AddItem(item *FlexItem) {
	fc.Items = append(fc.Items, item)
}

// Layout performs the complete flexbox layout algorithm
func (fc *FlexContainer) Layout() {
	if len(fc.Items) == 0 {
		return
	}

	// Step 1: Determine main and cross axis
	isRow := fc.Direction == "row" || fc.Direction == "row-reverse"
	isReverse := fc.Direction == "row-reverse" || fc.Direction == "column-reverse"

	var mainSize, crossSize float64
	if isRow {
		mainSize = fc.Width
		crossSize = fc.Height
	} else {
		mainSize = fc.Height
		crossSize = fc.Width
	}

	// Step 2: Collect items into flex lines
	fc.Lines = fc.collectIntoLines(mainSize, isRow)

	// Step 3: Resolve flexible lengths (grow/shrink)
	for _, line := range fc.Lines {
		fc.resolveFlexibleLengths(line, mainSize, isRow)
	}

	// Step 4: Align items on main axis
	for _, line := range fc.Lines {
		fc.alignMainAxis(line, mainSize, isReverse)
	}

	// Step 5: Determine cross sizes
	fc.determineCrossSizes(crossSize)

	// Step 6: Align items on cross axis
	fc.alignCrossAxis(crossSize)
}

// collectIntoLines divides items into flex lines based on wrap
func (fc *FlexContainer) collectIntoLines(mainSize float64, isRow bool) []FlexLine {
	if fc.Wrap == "nowrap" || fc.Wrap == "" {
		// All items in one line
		line := FlexLine{Items: fc.Items}
		for _, item := range fc.Items {
			if isRow {
				line.MainSize += item.FlexBaseSize + item.MarginMainStart + item.MarginMainEnd
			} else {
				line.MainSize += item.FlexBaseSize + item.MarginMainStart + item.MarginMainEnd
			}
		}
		if len(fc.Items) > 1 {
			line.MainSize += fc.Gap * float64(len(fc.Items)-1)
		}
		return []FlexLine{line}
	}

	// Wrap: distribute items into multiple lines
	var lines []FlexLine
	var currentLine FlexLine
	currentMainSize := float64(0)

	for _, item := range fc.Items {
		itemSize := item.FlexBaseSize + item.MarginMainStart + item.MarginMainEnd
		if len(currentLine.Items) > 0 {
			itemSize += fc.Gap
		}

		if len(currentLine.Items) > 0 && currentMainSize+itemSize > mainSize {
			// Start new line
			currentLine.MainSize = currentMainSize
			lines = append(lines, currentLine)
			currentLine = FlexLine{}
			currentMainSize = item.FlexBaseSize + item.MarginMainStart + item.MarginMainEnd
		} else {
			currentMainSize += itemSize
		}

		currentLine.Items = append(currentLine.Items, item)
	}

	if len(currentLine.Items) > 0 {
		currentLine.MainSize = currentMainSize
		lines = append(lines, currentLine)
	}

	return lines
}

// resolveFlexibleLengths implements the flex grow/shrink algorithm
func (fc *FlexContainer) resolveFlexibleLengths(line FlexLine, availableMain float64, isRow bool) {
	// Calculate free space
	usedSpace := float64(0)
	for _, item := range line.Items {
		usedSpace += item.FlexBaseSize + item.MarginMainStart + item.MarginMainEnd
	}
	if len(line.Items) > 1 {
		usedSpace += fc.Gap * float64(len(line.Items)-1)
	}

	freeSpace := availableMain - usedSpace

	if freeSpace > 0 {
		// Growing
		totalGrow := float64(0)
		for _, item := range line.Items {
			totalGrow += item.FlexGrow
		}

		if totalGrow > 0 {
			for _, item := range line.Items {
				item.FinalMainSize = item.FlexBaseSize + (freeSpace * item.FlexGrow / totalGrow)
			}
		} else {
			for _, item := range line.Items {
				item.FinalMainSize = item.FlexBaseSize
			}
		}
	} else if freeSpace < 0 {
		// Shrinking
		totalShrink := float64(0)
		for _, item := range line.Items {
			totalShrink += item.FlexShrink * item.FlexBaseSize
		}

		if totalShrink > 0 {
			for _, item := range line.Items {
				shrinkRatio := (item.FlexShrink * item.FlexBaseSize) / totalShrink
				item.FinalMainSize = item.FlexBaseSize + (freeSpace * shrinkRatio)
				if item.FinalMainSize < 0 {
					item.FinalMainSize = 0
				}
			}
		} else {
			for _, item := range line.Items {
				item.FinalMainSize = item.FlexBaseSize
			}
		}
	} else {
		for _, item := range line.Items {
			item.FinalMainSize = item.FlexBaseSize
		}
	}
}

// alignMainAxis positions items along main axis based on justify-content
func (fc *FlexContainer) alignMainAxis(line FlexLine, mainSize float64, isReverse bool) {
	// Calculate total used space
	usedSpace := float64(0)
	for _, item := range line.Items {
		usedSpace += item.FinalMainSize + item.MarginMainStart + item.MarginMainEnd
	}
	if len(line.Items) > 1 {
		usedSpace += fc.Gap * float64(len(line.Items)-1)
	}

	freeSpace := mainSize - usedSpace
	if freeSpace < 0 {
		freeSpace = 0
	}

	// Determine starting position and spacing
	var startPos, spacing float64
	numItems := len(line.Items)

	switch fc.JustifyContent {
	case "flex-start", "start", "":
		startPos = 0
		spacing = 0
	case "flex-end", "end":
		startPos = freeSpace
		spacing = 0
	case "center":
		startPos = freeSpace / 2
		spacing = 0
	case "space-between":
		startPos = 0
		if numItems > 1 {
			spacing = freeSpace / float64(numItems-1)
		}
	case "space-around":
		spacing = freeSpace / float64(numItems)
		startPos = spacing / 2
	case "space-evenly":
		spacing = freeSpace / float64(numItems+1)
		startPos = spacing
	}

	// Position items
	pos := startPos
	items := line.Items
	if isReverse {
		// Reverse order
		reversed := make([]*FlexItem, len(items))
		for i, item := range items {
			reversed[len(items)-1-i] = item
		}
		items = reversed
	}

	for i, item := range items {
		item.MainPos = pos + item.MarginMainStart
		pos += item.MarginMainStart + item.FinalMainSize + item.MarginMainEnd
		if i < numItems-1 {
			pos += fc.Gap + spacing
		}
	}
}

// determineCrossSizes calculates cross sizes for each line and item
func (fc *FlexContainer) determineCrossSizes(availableCross float64) {
	// For each line, find max cross size
	for i := range fc.Lines {
		maxCross := float64(0)
		for _, item := range fc.Lines[i].Items {
			itemCross := item.CrossSize + item.MarginCrossStart + item.MarginCrossEnd
			if itemCross > maxCross {
				maxCross = itemCross
			}
		}
		fc.Lines[i].CrossSize = maxCross
	}

	// If stretch, expand items to line cross size
	for i := range fc.Lines {
		for _, item := range fc.Lines[i].Items {
			align := item.AlignSelf
			if align == "" || align == "auto" {
				align = fc.AlignItems
			}

			if align == "stretch" {
				// Stretch to fill line
				item.FinalCrossSize = fc.Lines[i].CrossSize - item.MarginCrossStart - item.MarginCrossEnd
			} else {
				item.FinalCrossSize = item.CrossSize
			}
		}
	}
}

// alignCrossAxis positions items along cross axis based on align-items
func (fc *FlexContainer) alignCrossAxis(availableCross float64) {
	// Calculate total lines cross size
	totalCross := float64(0)
	for _, line := range fc.Lines {
		totalCross += line.CrossSize
	}
	if len(fc.Lines) > 1 {
		totalCross += fc.Gap * float64(len(fc.Lines)-1)
	}

	// Position lines based on align-content
	freeCross := availableCross - totalCross
	if freeCross < 0 {
		freeCross = 0
	}

	var lineStart, lineSpacing float64
	numLines := len(fc.Lines)

	switch fc.AlignContent {
	case "flex-start", "start", "":
		lineStart = 0
	case "flex-end", "end":
		lineStart = freeCross
	case "center":
		lineStart = freeCross / 2
	case "space-between":
		if numLines > 1 {
			lineSpacing = freeCross / float64(numLines-1)
		}
	case "space-around":
		lineSpacing = freeCross / float64(numLines)
		lineStart = lineSpacing / 2
	case "stretch":
		// Distribute free space to lines
		if numLines > 0 {
			extra := freeCross / float64(numLines)
			for i := range fc.Lines {
				fc.Lines[i].CrossSize += extra
			}
		}
	}

	// Position lines
	crossPos := lineStart
	for i := range fc.Lines {
		fc.Lines[i].CrossStart = crossPos
		crossPos += fc.Lines[i].CrossSize + fc.Gap + lineSpacing
	}

	// Position items within lines based on align-items
	for _, line := range fc.Lines {
		for _, item := range line.Items {
			align := item.AlignSelf
			if align == "" || align == "auto" {
				align = fc.AlignItems
			}

			lineSize := line.CrossSize
			itemSize := item.FinalCrossSize + item.MarginCrossStart + item.MarginCrossEnd

			switch align {
			case "flex-start", "start":
				item.CrossPos = line.CrossStart + item.MarginCrossStart
			case "flex-end", "end":
				item.CrossPos = line.CrossStart + lineSize - itemSize + item.MarginCrossStart
			case "center":
				item.CrossPos = line.CrossStart + (lineSize-itemSize)/2 + item.MarginCrossStart
			case "stretch", "":
				item.CrossPos = line.CrossStart + item.MarginCrossStart
			case "baseline":
				// Simplified: treat as flex-start
				item.CrossPos = line.CrossStart + item.MarginCrossStart
			}
		}
	}
}

// GetItemPosition returns the final position and size for an item (main, cross order)
func (fc *FlexContainer) GetItemPosition(item *FlexItem) (x, y, width, height float64) {
	isRow := fc.Direction == "row" || fc.Direction == "row-reverse"

	if isRow {
		return item.MainPos, item.CrossPos, item.FinalMainSize, item.FinalCrossSize
	}
	return item.CrossPos, item.MainPos, item.FinalCrossSize, item.FinalMainSize
}
