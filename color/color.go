package color

import (
	"github.com/fatih/color"
)

// Success writes string for success messages.
var Success = color.New(color.FgHiGreen).SprintFunc()

// Info writes string for informational messages.
var Info = color.New(color.FgHiCyan).SprintFunc()

// Warn writes string for warning messages.
var Warn = color.New(color.FgHiYellow).SprintFunc()

// Error writes string for error messages.
var Error = color.New(color.FgHiRed).SprintFunc()

// Mute writes string for muted messages.
var Mute = color.New(color.FgHiBlack).SprintFunc()

// Important writes string for important messages.
var Important = color.New(color.BgRed, color.Bold).SprintFunc()

// Private writes string for private messages.
var Private = color.New(color.Underline).SprintFunc()
