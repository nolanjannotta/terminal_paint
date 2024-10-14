package main

import "github.com/charmbracelet/lipgloss"

type params struct {
	tip   int
	color int
	move  lipgloss.Style
	erase lipgloss.Style
}

type offset struct {
	x, y, startingX, startingY int
}

type tip struct {
	char string
	x    int
}

type model struct {
	width, height int
	params        params
	offset        offset
	isDrawing     bool
	colors        []int
	tips          []tip
	canvas        lipgloss.Style
	pixelMap      map[[2]int]lipgloss.Style
}
