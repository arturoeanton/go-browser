// Package values provides CSS value types for the Gocko CSS engine
// Color implementation follows CSS Color Level 4 specification
package values

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

// =============================================================================
// COLOR VALUES
// CSS Colors: https://www.w3.org/TR/css-color-4/
// =============================================================================

// Color represents a CSS color value with RGBA components
type Color struct {
	R, G, B, A uint8
}

// Transparent returns a transparent color
func Transparent() Color {
	return Color{0, 0, 0, 0}
}

// Black returns black
func Black() Color {
	return Color{0, 0, 0, 255}
}

// White returns white
func White() Color {
	return Color{255, 255, 255, 255}
}

// RGB creates a color from RGB values
func RGB(r, g, b uint8) Color {
	return Color{r, g, b, 255}
}

// RGBA creates a color from RGBA values
func RGBA(r, g, b, a uint8) Color {
	return Color{r, g, b, a}
}

// FromHex creates a color from a hex string (#RGB, #RGBA, #RRGGBB, #RRGGBBAA)
func FromHex(hex string) (Color, error) {
	hex = strings.TrimPrefix(hex, "#")
	hex = strings.ToLower(hex)

	var r, g, b, a uint8 = 0, 0, 0, 255

	switch len(hex) {
	case 3: // #RGB
		r = hexDigit(hex[0]) * 17
		g = hexDigit(hex[1]) * 17
		b = hexDigit(hex[2]) * 17
	case 4: // #RGBA
		r = hexDigit(hex[0]) * 17
		g = hexDigit(hex[1]) * 17
		b = hexDigit(hex[2]) * 17
		a = hexDigit(hex[3]) * 17
	case 6: // #RRGGBB
		r = hexByte(hex[0:2])
		g = hexByte(hex[2:4])
		b = hexByte(hex[4:6])
	case 8: // #RRGGBBAA
		r = hexByte(hex[0:2])
		g = hexByte(hex[2:4])
		b = hexByte(hex[4:6])
		a = hexByte(hex[6:8])
	default:
		return Transparent(), fmt.Errorf("invalid hex color: #%s", hex)
	}

	return Color{r, g, b, a}, nil
}

func hexDigit(c byte) uint8 {
	if c >= '0' && c <= '9' {
		return c - '0'
	}
	if c >= 'a' && c <= 'f' {
		return c - 'a' + 10
	}
	return 0
}

func hexByte(s string) uint8 {
	v, _ := strconv.ParseUint(s, 16, 8)
	return uint8(v)
}

// ToRGBA converts to Go's standard color.RGBA
func (c Color) ToRGBA() color.RGBA {
	return color.RGBA{c.R, c.G, c.B, c.A}
}

// IsTransparent returns true if color is fully transparent
func (c Color) IsTransparent() bool {
	return c.A == 0
}

// IsOpaque returns true if color is fully opaque
func (c Color) IsOpaque() bool {
	return c.A == 255
}

// WithAlpha returns a new color with modified alpha
func (c Color) WithAlpha(a uint8) Color {
	return Color{c.R, c.G, c.B, a}
}

// Blend blends two colors based on alpha of the top color
func (c Color) Blend(top Color) Color {
	if top.A == 255 {
		return top
	}
	if top.A == 0 {
		return c
	}

	alpha := float64(top.A) / 255
	return Color{
		R: uint8(float64(c.R)*(1-alpha) + float64(top.R)*alpha),
		G: uint8(float64(c.G)*(1-alpha) + float64(top.G)*alpha),
		B: uint8(float64(c.B)*(1-alpha) + float64(top.B)*alpha),
		A: uint8(float64(c.A)*(1-alpha) + float64(top.A)),
	}
}

// String returns a CSS representation of the color
func (c Color) String() string {
	if c.A == 255 {
		return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
	}
	return fmt.Sprintf("rgba(%d, %d, %d, %.2f)", c.R, c.G, c.B, float64(c.A)/255)
}

// =============================================================================
// CSS NAMED COLORS (CSS Color Level 4)
// =============================================================================

var namedColors = map[string]Color{
	// Basic colors
	"transparent": {0, 0, 0, 0},
	"black":       {0, 0, 0, 255},
	"white":       {255, 255, 255, 255},
	"red":         {255, 0, 0, 255},
	"green":       {0, 128, 0, 255},
	"blue":        {0, 0, 255, 255},
	"yellow":      {255, 255, 0, 255},
	"cyan":        {0, 255, 255, 255},
	"magenta":     {255, 0, 255, 255},

	// Extended colors
	"gray":      {128, 128, 128, 255},
	"grey":      {128, 128, 128, 255},
	"silver":    {192, 192, 192, 255},
	"maroon":    {128, 0, 0, 255},
	"olive":     {128, 128, 0, 255},
	"lime":      {0, 255, 0, 255},
	"aqua":      {0, 255, 255, 255},
	"teal":      {0, 128, 128, 255},
	"navy":      {0, 0, 128, 255},
	"fuchsia":   {255, 0, 255, 255},
	"purple":    {128, 0, 128, 255},
	"orange":    {255, 165, 0, 255},
	"pink":      {255, 192, 203, 255},
	"brown":     {165, 42, 42, 255},
	"coral":     {255, 127, 80, 255},
	"crimson":   {220, 20, 60, 255},
	"gold":      {255, 215, 0, 255},
	"indigo":    {75, 0, 130, 255},
	"khaki":     {240, 230, 140, 255},
	"lavender":  {230, 230, 250, 255},
	"salmon":    {250, 128, 114, 255},
	"skyblue":   {135, 206, 235, 255},
	"slategray": {112, 128, 144, 255},
	"steelblue": {70, 130, 180, 255},
	"tomato":    {255, 99, 71, 255},
	"turquoise": {64, 224, 208, 255},
	"violet":    {238, 130, 238, 255},
	"wheat":     {245, 222, 179, 255},

	// Gray scale
	"dimgray":    {105, 105, 105, 255},
	"darkgray":   {169, 169, 169, 255},
	"lightgray":  {211, 211, 211, 255},
	"gainsboro":  {220, 220, 220, 255},
	"whitesmoke": {245, 245, 245, 255},

	// Additional modern colors
	"aliceblue":         {240, 248, 255, 255},
	"antiquewhite":      {250, 235, 215, 255},
	"azure":             {240, 255, 255, 255},
	"beige":             {245, 245, 220, 255},
	"bisque":            {255, 228, 196, 255},
	"blanchedalmond":    {255, 235, 205, 255},
	"blueviolet":        {138, 43, 226, 255},
	"burlywood":         {222, 184, 135, 255},
	"cadetblue":         {95, 158, 160, 255},
	"chartreuse":        {127, 255, 0, 255},
	"chocolate":         {210, 105, 30, 255},
	"cornflowerblue":    {100, 149, 237, 255},
	"cornsilk":          {255, 248, 220, 255},
	"darkblue":          {0, 0, 139, 255},
	"darkcyan":          {0, 139, 139, 255},
	"darkgoldenrod":     {184, 134, 11, 255},
	"darkgreen":         {0, 100, 0, 255},
	"darkkhaki":         {189, 183, 107, 255},
	"darkmagenta":       {139, 0, 139, 255},
	"darkolivegreen":    {85, 107, 47, 255},
	"darkorange":        {255, 140, 0, 255},
	"darkorchid":        {153, 50, 204, 255},
	"darkred":           {139, 0, 0, 255},
	"darksalmon":        {233, 150, 122, 255},
	"darkseagreen":      {143, 188, 143, 255},
	"darkslateblue":     {72, 61, 139, 255},
	"darkslategray":     {47, 79, 79, 255},
	"darkturquoise":     {0, 206, 209, 255},
	"darkviolet":        {148, 0, 211, 255},
	"deeppink":          {255, 20, 147, 255},
	"deepskyblue":       {0, 191, 255, 255},
	"dodgerblue":        {30, 144, 255, 255},
	"firebrick":         {178, 34, 34, 255},
	"floralwhite":       {255, 250, 240, 255},
	"forestgreen":       {34, 139, 34, 255},
	"ghostwhite":        {248, 248, 255, 255},
	"goldenrod":         {218, 165, 32, 255},
	"greenyellow":       {173, 255, 47, 255},
	"honeydew":          {240, 255, 240, 255},
	"hotpink":           {255, 105, 180, 255},
	"indianred":         {205, 92, 92, 255},
	"ivory":             {255, 255, 240, 255},
	"lawngreen":         {124, 252, 0, 255},
	"lemonchiffon":      {255, 250, 205, 255},
	"lightblue":         {173, 216, 230, 255},
	"lightcoral":        {240, 128, 128, 255},
	"lightcyan":         {224, 255, 255, 255},
	"lightgreen":        {144, 238, 144, 255},
	"lightpink":         {255, 182, 193, 255},
	"lightsalmon":       {255, 160, 122, 255},
	"lightseagreen":     {32, 178, 170, 255},
	"lightskyblue":      {135, 206, 250, 255},
	"lightslategray":    {119, 136, 153, 255},
	"lightsteelblue":    {176, 196, 222, 255},
	"lightyellow":       {255, 255, 224, 255},
	"limegreen":         {50, 205, 50, 255},
	"linen":             {250, 240, 230, 255},
	"mediumaquamarine":  {102, 205, 170, 255},
	"mediumblue":        {0, 0, 205, 255},
	"mediumorchid":      {186, 85, 211, 255},
	"mediumpurple":      {147, 112, 219, 255},
	"mediumseagreen":    {60, 179, 113, 255},
	"mediumslateblue":   {123, 104, 238, 255},
	"mediumspringgreen": {0, 250, 154, 255},
	"mediumturquoise":   {72, 209, 204, 255},
	"mediumvioletred":   {199, 21, 133, 255},
	"midnightblue":      {25, 25, 112, 255},
	"mintcream":         {245, 255, 250, 255},
	"mistyrose":         {255, 228, 225, 255},
	"moccasin":          {255, 228, 181, 255},
	"navajowhite":       {255, 222, 173, 255},
	"oldlace":           {253, 245, 230, 255},
	"olivedrab":         {107, 142, 35, 255},
	"orangered":         {255, 69, 0, 255},
	"orchid":            {218, 112, 214, 255},
	"palegoldenrod":     {238, 232, 170, 255},
	"palegreen":         {152, 251, 152, 255},
	"paleturquoise":     {175, 238, 238, 255},
	"palevioletred":     {219, 112, 147, 255},
	"papayawhip":        {255, 239, 213, 255},
	"peachpuff":         {255, 218, 185, 255},
	"peru":              {205, 133, 63, 255},
	"plum":              {221, 160, 221, 255},
	"powderblue":        {176, 224, 230, 255},
	"rosybrown":         {188, 143, 143, 255},
	"royalblue":         {65, 105, 225, 255},
	"saddlebrown":       {139, 69, 19, 255},
	"sandybrown":        {244, 164, 96, 255},
	"seagreen":          {46, 139, 87, 255},
	"seashell":          {255, 245, 238, 255},
	"sienna":            {160, 82, 45, 255},
	"slateblue":         {106, 90, 205, 255},
	"snow":              {255, 250, 250, 255},
	"springgreen":       {0, 255, 127, 255},
	"tan":               {210, 180, 140, 255},
	"thistle":           {216, 191, 216, 255},
	"yellowgreen":       {154, 205, 50, 255},
}

// ParseColor parses a CSS color string
func ParseColor(s string) (Color, error) {
	s = strings.TrimSpace(strings.ToLower(s))

	// Check named colors first
	if c, ok := namedColors[s]; ok {
		return c, nil
	}

	// Hex color
	if strings.HasPrefix(s, "#") {
		return FromHex(s)
	}

	// rgb() / rgba()
	if strings.HasPrefix(s, "rgb") {
		return parseRGBFunction(s)
	}

	// hsl() / hsla()
	if strings.HasPrefix(s, "hsl") {
		return parseHSLFunction(s)
	}

	return Transparent(), fmt.Errorf("invalid color: %s", s)
}

func parseRGBFunction(s string) (Color, error) {
	// rgb(255, 0, 0) or rgba(255, 0, 0, 0.5) or rgb(255 0 0 / 50%)
	s = strings.TrimPrefix(s, "rgba(")
	s = strings.TrimPrefix(s, "rgb(")
	s = strings.TrimSuffix(s, ")")

	// Handle modern syntax with / for alpha
	var alpha float64 = 1.0
	if idx := strings.Index(s, "/"); idx != -1 {
		alphaStr := strings.TrimSpace(s[idx+1:])
		s = s[:idx]
		if strings.HasSuffix(alphaStr, "%") {
			alphaStr = strings.TrimSuffix(alphaStr, "%")
			a, _ := strconv.ParseFloat(alphaStr, 64)
			alpha = a / 100
		} else {
			alpha, _ = strconv.ParseFloat(alphaStr, 64)
		}
	}

	// Parse components
	s = strings.ReplaceAll(s, ",", " ")
	parts := strings.Fields(s)

	if len(parts) < 3 {
		return Transparent(), fmt.Errorf("invalid rgb: not enough components")
	}

	r := parseColorComponent(parts[0])
	g := parseColorComponent(parts[1])
	b := parseColorComponent(parts[2])

	if len(parts) >= 4 {
		alpha, _ = strconv.ParseFloat(parts[3], 64)
	}

	return Color{r, g, b, uint8(alpha * 255)}, nil
}

func parseColorComponent(s string) uint8 {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "%") {
		s = strings.TrimSuffix(s, "%")
		v, _ := strconv.ParseFloat(s, 64)
		return uint8(v * 255 / 100)
	}
	v, _ := strconv.ParseFloat(s, 64)
	if v > 255 {
		v = 255
	}
	if v < 0 {
		v = 0
	}
	return uint8(v)
}

func parseHSLFunction(s string) (Color, error) {
	// hsl(180, 50%, 50%) or hsla(180, 50%, 50%, 0.5)
	s = strings.TrimPrefix(s, "hsla(")
	s = strings.TrimPrefix(s, "hsl(")
	s = strings.TrimSuffix(s, ")")

	// Handle modern syntax with / for alpha
	var alpha float64 = 1.0
	if idx := strings.Index(s, "/"); idx != -1 {
		alphaStr := strings.TrimSpace(s[idx+1:])
		s = s[:idx]
		if strings.HasSuffix(alphaStr, "%") {
			alphaStr = strings.TrimSuffix(alphaStr, "%")
			a, _ := strconv.ParseFloat(alphaStr, 64)
			alpha = a / 100
		} else {
			alpha, _ = strconv.ParseFloat(alphaStr, 64)
		}
	}

	s = strings.ReplaceAll(s, ",", " ")
	parts := strings.Fields(s)

	if len(parts) < 3 {
		return Transparent(), fmt.Errorf("invalid hsl: not enough components")
	}

	// Parse H (degrees)
	hStr := strings.TrimSuffix(parts[0], "deg")
	h, _ := strconv.ParseFloat(hStr, 64)
	h = h / 360 // Normalize to 0-1

	// Parse S and L (percentages)
	sStr := strings.TrimSuffix(parts[1], "%")
	sat, _ := strconv.ParseFloat(sStr, 64)
	sat = sat / 100

	lStr := strings.TrimSuffix(parts[2], "%")
	light, _ := strconv.ParseFloat(lStr, 64)
	light = light / 100

	if len(parts) >= 4 {
		alpha, _ = strconv.ParseFloat(parts[3], 64)
	}

	// Convert HSL to RGB
	r, g, b := hslToRGB(h, sat, light)

	return Color{r, g, b, uint8(alpha * 255)}, nil
}

// hslToRGB converts HSL to RGB
func hslToRGB(h, s, l float64) (uint8, uint8, uint8) {
	if s == 0 {
		v := uint8(l * 255)
		return v, v, v
	}

	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q

	r := hueToRGB(p, q, h+1.0/3.0)
	g := hueToRGB(p, q, h)
	b := hueToRGB(p, q, h-1.0/3.0)

	return uint8(r * 255), uint8(g * 255), uint8(b * 255)
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 0.5 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}
