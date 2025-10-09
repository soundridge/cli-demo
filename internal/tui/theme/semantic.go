package theme

// SemanticColors provides a modern, semantic color organization
// similar to the TypeScript implementation's semantic-colors structure.
type SemanticColors struct {
	Text       TextColors
	Background BackgroundColors
	Border     BorderColors
	UI         UIColors
	Status     StatusColors
}

// ToSemanticColors converts a ColorsTheme to SemanticColors for easier access.
func (t ColorsTheme) ToSemanticColors() SemanticColors {
	return SemanticColors{
		Text: TextColors{
			Primary:   t.Foreground,
			Secondary: t.Gray,
			Link:      t.AccentBlue,
			Accent:    t.AccentPurple,
		},
		Background: BackgroundColors{
			Primary: t.Background,
			Diff: &DiffColors{
				Added:   t.DiffAdded,
				Removed: t.DiffRemoved,
			},
		},
		Border: BorderColors{
			Default: t.Gray,
			Focused: t.AccentBlue,
		},
		UI: UIColors{
			Comment:  t.Comment,
			Symbol:   t.AccentCyan,
			Gradient: t.GradientColors,
		},
		Status: StatusColors{
			Error:   t.AccentRed,
			Success: t.AccentGreen,
			Warning: t.AccentYellow,
		},
	}
}
