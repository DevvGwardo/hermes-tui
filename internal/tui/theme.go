package tui

import "github.com/charmbracelet/lipgloss"

// ThemeName identifies a color theme.
type ThemeName string

const (
	ThemeOcean    ThemeName = "ocean"
	ThemeAmber    ThemeName = "amber"
	ThemeRose     ThemeName = "rose"
	ThemeForest   ThemeName = "forest"
	ThemeAquarium ThemeName = "aquarium"
)

// ThemeNames lists all available themes.
var ThemeNames = []ThemeName{ThemeOcean, ThemeAmber, ThemeRose, ThemeForest, ThemeAquarium}

// Palette defines the colors for a theme.
type Palette struct {
	Primary      lipgloss.Color
	Secondary    lipgloss.Color
	Accent       lipgloss.Color
	Bg           lipgloss.Color
	BgSubtle     lipgloss.Color
	Fg           lipgloss.Color
	FgMuted      lipgloss.Color
	Success      lipgloss.Color
	Warning      lipgloss.Color
	Error        lipgloss.Color
	UserBg       lipgloss.Color
	UserBorder   lipgloss.Color
	AssistBg     lipgloss.Color
	AssistBorder lipgloss.Color
	HeaderFg     lipgloss.Color
	HeaderBg     lipgloss.Color
	StatusFg     lipgloss.Color
	StatusBg     lipgloss.Color
	InputFg      lipgloss.Color
	InputBg      lipgloss.Color
	Muted        lipgloss.Color
}

var palettes = map[ThemeName]Palette{
	ThemeOcean: {
		Primary:      lipgloss.Color("#5EBED6"),
		Secondary:    lipgloss.Color("#F0A870"),
		Accent:       lipgloss.Color("#7DD3FC"),
		Bg:           lipgloss.Color("#1A1B2E"),
		BgSubtle:     lipgloss.Color("#252640"),
		Fg:           lipgloss.Color("#E8E8F0"),
		FgMuted:      lipgloss.Color("#8888A0"),
		Success:      lipgloss.Color("#6CC890"),
		Warning:      lipgloss.Color("#F6C453"),
		Error:        lipgloss.Color("#E87070"),
		UserBg:       lipgloss.Color("#1E2A3A"),
		UserBorder:   lipgloss.Color("#F0A870"),
		AssistBg:     lipgloss.Color("#1A1B2E"),
		AssistBorder: lipgloss.Color("#5EBED6"),
		HeaderFg:     lipgloss.Color("#5EBED6"),
		HeaderBg:     lipgloss.Color("#0F0F1A"),
		StatusFg:     lipgloss.Color("#8888A0"),
		StatusBg:     lipgloss.Color("#0F0F1A"),
		InputFg:      lipgloss.Color("#E8E8F0"),
		InputBg:      lipgloss.Color("#252640"),
		Muted:        lipgloss.Color("#4A4A6A"),
	},
	ThemeAmber: {
		Primary:      lipgloss.Color("#F6C453"),
		Secondary:    lipgloss.Color("#F2A65A"),
		Accent:       lipgloss.Color("#FFD580"),
		Bg:           lipgloss.Color("#1C1A16"),
		BgSubtle:     lipgloss.Color("#2A2720"),
		Fg:           lipgloss.Color("#F0E8D8"),
		FgMuted:      lipgloss.Color("#A09880"),
		Success:      lipgloss.Color("#6CC890"),
		Warning:      lipgloss.Color("#F6C453"),
		Error:        lipgloss.Color("#E87070"),
		UserBg:       lipgloss.Color("#2A2518"),
		UserBorder:   lipgloss.Color("#F2A65A"),
		AssistBg:     lipgloss.Color("#1C1A16"),
		AssistBorder: lipgloss.Color("#F6C453"),
		HeaderFg:     lipgloss.Color("#F6C453"),
		HeaderBg:     lipgloss.Color("#141210"),
		StatusFg:     lipgloss.Color("#A09880"),
		StatusBg:     lipgloss.Color("#141210"),
		InputFg:      lipgloss.Color("#F0E8D8"),
		InputBg:      lipgloss.Color("#2A2720"),
		Muted:        lipgloss.Color("#4A4838"),
	},
	ThemeRose: {
		Primary:      lipgloss.Color("#E87CA0"),
		Secondary:    lipgloss.Color("#C8A0E0"),
		Accent:       lipgloss.Color("#F0A0C0"),
		Bg:           lipgloss.Color("#1C1A20"),
		BgSubtle:     lipgloss.Color("#2A2530"),
		Fg:           lipgloss.Color("#F0E8F0"),
		FgMuted:      lipgloss.Color("#9888A0"),
		Success:      lipgloss.Color("#6CC890"),
		Warning:      lipgloss.Color("#F6C453"),
		Error:        lipgloss.Color("#E87070"),
		UserBg:       lipgloss.Color("#2A1825"),
		UserBorder:   lipgloss.Color("#E87CA0"),
		AssistBg:     lipgloss.Color("#1C1A20"),
		AssistBorder: lipgloss.Color("#C8A0E0"),
		HeaderFg:     lipgloss.Color("#E87CA0"),
		HeaderBg:     lipgloss.Color("#121018"),
		StatusFg:     lipgloss.Color("#9888A0"),
		StatusBg:     lipgloss.Color("#121018"),
		InputFg:      lipgloss.Color("#F0E8F0"),
		InputBg:      lipgloss.Color("#2A2530"),
		Muted:        lipgloss.Color("#4A4050"),
	},
	ThemeForest: {
		Primary:      lipgloss.Color("#6CC890"),
		Secondary:    lipgloss.Color("#D8B870"),
		Accent:       lipgloss.Color("#90E0B0"),
		Bg:           lipgloss.Color("#161C18"),
		BgSubtle:     lipgloss.Color("#202A22"),
		Fg:           lipgloss.Color("#E0F0E0"),
		FgMuted:      lipgloss.Color("#80A088"),
		Success:      lipgloss.Color("#6CC890"),
		Warning:      lipgloss.Color("#F6C453"),
		Error:        lipgloss.Color("#E87070"),
		UserBg:       lipgloss.Color("#182A1E"),
		UserBorder:   lipgloss.Color("#D8B870"),
		AssistBg:     lipgloss.Color("#161C18"),
		AssistBorder: lipgloss.Color("#6CC890"),
		HeaderFg:     lipgloss.Color("#6CC890"),
		HeaderBg:     lipgloss.Color("#0E120E"),
		StatusFg:     lipgloss.Color("#80A088"),
		StatusBg:     lipgloss.Color("#0E120E"),
		InputFg:      lipgloss.Color("#E0F0E0"),
		InputBg:      lipgloss.Color("#202A22"),
		Muted:        lipgloss.Color("#384840"),
	},
	ThemeAquarium: {
		Primary:      lipgloss.Color("#00B4D8"),
		Secondary:    lipgloss.Color("#F4A261"),
		Accent:       lipgloss.Color("#48CAE4"),
		Bg:           lipgloss.Color("#0A1628"),
		BgSubtle:     lipgloss.Color("#0F2035"),
		Fg:           lipgloss.Color("#CAF0F8"),
		FgMuted:      lipgloss.Color("#5E8BA0"),
		Success:      lipgloss.Color("#2EC4B6"),
		Warning:      lipgloss.Color("#F4A261"),
		Error:        lipgloss.Color("#E76F51"),
		UserBg:       lipgloss.Color("#0D1F30"),
		UserBorder:   lipgloss.Color("#F4A261"),
		AssistBg:     lipgloss.Color("#0A1628"),
		AssistBorder: lipgloss.Color("#00B4D8"),
		HeaderFg:     lipgloss.Color("#00B4D8"),
		HeaderBg:     lipgloss.Color("#060D18"),
		StatusFg:     lipgloss.Color("#5E8BA0"),
		StatusBg:     lipgloss.Color("#060D18"),
		InputFg:      lipgloss.Color("#CAF0F8"),
		InputBg:      lipgloss.Color("#0F2035"),
		Muted:        lipgloss.Color("#1A3040"),
	},
}

// Theme holds a named palette.
type Theme struct {
	Name    ThemeName
	Palette Palette
}

// NewTheme creates a theme by name, defaulting to ocean.
func NewTheme(name string) Theme {
	tn := ThemeName(name)
	if _, ok := palettes[tn]; !ok {
		tn = ThemeOcean
	}
	return Theme{Name: tn, Palette: palettes[tn]}
}
