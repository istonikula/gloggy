// Package appshell provides the top-level layout and global commands.
package appshell

// KeyBinding represents a single keybinding with its key and description.
type KeyBinding struct {
	Key  string
	Desc string
}

// Domain represents a logical group of keybindings.
type Domain string

const (
	DomainEntryList   Domain = "entry-list"
	DomainDetailPane  Domain = "detail-pane"
	DomainFilterPanel Domain = "filter-panel"
	DomainGlobal      Domain = "global"
)

// KeybindingRegistry maps domains to their keybindings.
type KeybindingRegistry map[Domain][]KeyBinding

// DefaultKeybindings returns the full keybinding registry for all domains.
func DefaultKeybindings() KeybindingRegistry {
	return KeybindingRegistry{
		DomainEntryList: {
			{Key: "j/↓", Desc: "Move cursor down"},
			{Key: "k/↑", Desc: "Move cursor up"},
			{Key: "g", Desc: "Go to top"},
			{Key: "G", Desc: "Go to bottom"},
			{Key: "Ctrl+d", Desc: "Half page down"},
			{Key: "Ctrl+u", Desc: "Half page up"},
			{Key: "Enter", Desc: "Open entry in detail pane"},
			{Key: "/", Desc: "Search within list (focus-scoped — detail pane has its own / search)"},
			{Key: "m", Desc: "Toggle mark on current entry"},
			{Key: "u", Desc: "Jump to next mark"},
			{Key: "U", Desc: "Jump to previous mark"},
			{Key: "M", Desc: "Clear all marks"},
		},
		DomainDetailPane: {
			{Key: "j/↓", Desc: "Scroll down"},
			{Key: "k/↑", Desc: "Scroll up"},
			{Key: "q/Esc", Desc: "Close detail pane"},
			{Key: "/", Desc: "Search inside this pane (Enter commits to navigate mode)"},
			{Key: "n/N", Desc: "Next/prev search match"},
		},
		DomainFilterPanel: {
			{Key: "Enter", Desc: "Commit filter"},
			{Key: "Esc", Desc: "Cancel filter input"},
			{Key: "Tab", Desc: "Cycle between field selector and pattern"},
			{Key: "Space", Desc: "Toggle filter enabled"},
			{Key: "d", Desc: "Delete filter"},
		},
		DomainGlobal: {
			{Key: "?", Desc: "Toggle help overlay"},
			{Key: "Tab", Desc: "Cycle focus between panels"},
			{Key: "f", Desc: "Open filter panel"},
			{Key: "y", Desc: "Copy marked entries to clipboard"},
			{Key: "T", Desc: "Open theme selector"},
			{Key: "q", Desc: "Quit"},
		},
	}
}

// Domains returns the ordered list of domains for display.
func Domains() []Domain {
	return []Domain{DomainGlobal, DomainEntryList, DomainDetailPane, DomainFilterPanel}
}
