package shared

var ColorList = []string{
	"#5470C6", // Blue
	"#3BA272", // Green
	"#FC8452", // Orange
	"#73C0DE", // Light blue
	"#EE6666", // Red
	"#FAC858", // Yellow
	"#9A60B4", // Purple
	"#EA7CCC", // Pink
	"#91CC75", // Lime
	"#FF9F7F", // Coral

	"#3E5A9E", // Navy Depth (Blue)
	"#2E7D32", // Forest Canopy (Green)
	"#EF6C00", // Burnt Sienna (Orange)
	"#7E57C2", // Amethyst Veil (Purple)
	"#F9A825", // Goldenrod Shine (Yellow)
	"#6A8ACF", // Steel Blue (Blue)
	"#4CAF50", // Emerald Leaf (Green)
	"#FF8F00", // Amber Flame (Orange)
	"#AB47BC", // Fuchsia Bloom (Pink)
	"#FFEB3B", // Lemon Zest (Yellow)

	"#2B4E72", // Midnight Slate (Blue)
	"#1B5E20", // Pine Shadow (Green)
	"#D84315", // Rust Ember (Red)
	"#512DA8", // Violet Shadow (Purple)
	"#F57F17", // Saffron Warmth (Yellow)
	"#4A90E2", // Cobalt Glow (Blue)
	"#66BB6A", // Verdant Bloom (Green)
	"#FF5722", // Coral Fire (Orange)
	"#BA68C8", // Orchid Haze (Purple)
	"#FFF176", // Banana Glow (Yellow)

	"#1E3A5F", // Deep Ocean (Blue)
	"#00695C", // Jade Depth (Green)
	"#BF360C", // Crimson Ember (Red)
	"#673AB7", // Plum Depth (Purple)
	"#C0CA33", // Chartreuse Edge (Lime)
	"#7FB3D5", // Azure Mist (Blue)
	"#81C784", // Moss Glow (Green)
	"#FFAB91", // Peach Sunset (Orange)
	"#E040FB", // Magenta Spark (Pink)
	"#DCE775", // Lime Radiance (Lime)

	"#335C8A", // Indigo Wave (Blue)
	"#388E3C", // Olive Ridge (Green)
	"#E64A19", // Tangerine Blaze (Orange)
	"#9575CD", // Lavender Dusk (Purple)
	"#78909C", // Slate Gray-Blue (Neutral)
	"#5C9EAD", // Teal Horizon (Blue-Green)
	"#AED581", // Sage Whisper (Green)
	"#FF7043", // Salmon Glow (Orange)
	"#F06292", // Rose Quartz (Pink)
	"#A1887F", // Taupe Earth (Neutral)
}

var (
	colorMap = make(map[string]int)
	i        int
)

func GetNextColorFor(key string) string {
	if _, has := colorMap[key]; has {
		return ColorList[colorMap[key]]
	}

	colorIndex := i % len(ColorList)
	color := ColorList[colorIndex]

	if i == len(ColorList) {
		i = 0
	} else {
		i++
	}

	return color
}
