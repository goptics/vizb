package shared

// Background is the bar category background: one wire object folding ECharts
// showBackground (Active) and backgroundStyle (the style fields). Active is
// the on-switch; every style field is optional and maps straight onto the
// rendered series. Numeric style fields are pointers so an explicit zero
// (e.g. borderWidth=0) survives the round trip.
type Background struct {
	Active        bool          `json:"active"`
	Color         string        `json:"color,omitempty"`
	BorderColor   string        `json:"borderColor,omitempty"`
	BorderWidth   *float64      `json:"borderWidth,omitempty"`
	BorderType    string        `json:"borderType,omitempty"`
	BorderRadius  *BorderRadius `json:"borderRadius,omitempty"`
	ShadowBlur    *float64      `json:"shadowBlur,omitempty"`
	ShadowColor   string        `json:"shadowColor,omitempty"`
	ShadowOffsetX *float64      `json:"shadowOffsetX,omitempty"`
	ShadowOffsetY *float64      `json:"shadowOffsetY,omitempty"`
	Opacity       *float64      `json:"opacity,omitempty"`
}
