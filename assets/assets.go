// Package assets embeds static assets like images for the devpipe CLI.
package assets

import _ "embed"

// MascotImage contains the embedded mascot PNG image
//
//go:embed squirrel-blank-eyes-transparent-cropped.png
var MascotImage []byte
