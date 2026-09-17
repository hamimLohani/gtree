// Package tui provides a Bubble Tea program that shows a spinner while
// scanning for git repositories, then switches to the styled tree view.
package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hamimlohani/gtree/internal/config"
	"github.com/hamimlohani/gtree/internal/scanner"
	"github.com/hamimlohani/gtree/internal/theme"
	"github.com/hamimlohani/gtree/internal/tree"
	"golang.org/x/term"
)

const watchInterval = 5 * time.Second

// ── Messages ──────────────────────────────────────────────────────────────────

type scanDoneMsg struct {
	repos []*scanner.Repo
	err   error
}

type watchTickMsg struct{}

// ── Model ─────────────────────────────────────────────────────────────────────

type Model struct {
	// config
	scanRoot string
	opts     scanner.Options
	cfg      *config.Config
	watch    bool

	// ui state
	spinner     spinner.Model
	scanning    bool
	repos       []*scanner.Repo
	scanErr     error
	output      string    // cached rendered tree
	lastUpdated time.Time // when last scan completed
	scanCount   int       // how many scans have run (for watch badge)
	width       int
}

func New(scanRoot string, opts scanner.Options, cfg *config.Config, watch bool) Model {
	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa"))

	width, _, _ := term.GetSize(1)
	if width <= 0 {
		width = 100
	}

	return Model{
		scanRoot: scanRoot,
		opts:     opts,
		cfg:      cfg,
		watch:    watch,
		spinner:  sp,
		scanning: true,
		width:    width,
	}
}

// ── Init ──────────────────────────────────────────────────────────────────────

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.doScan())
}

// ── Update ────────────────────────────────────────────────────────────────────

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "Q", "ctrl+c", "esc":
			return m, tea.Quit
		case "r", "R":
			// Manual refresh in watch mode.
			if m.watch && !m.scanning {
				m.scanning = true
				return m, tea.Batch(m.spinner.Tick, m.doScan())
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width

	case spinner.TickMsg:
		if m.scanning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case scanDoneMsg:
		m.scanning = false
		m.repos = msg.repos
		m.scanErr = msg.err
		m.lastUpdated = time.Now()
		m.scanCount++
		m.output = m.renderTree()

		if !m.watch {
			return m, tea.Quit
		}
		// Schedule next automatic refresh.
		return m, tea.Tick(watchInterval, func(time.Time) tea.Msg {
			return watchTickMsg{}
		})

	case watchTickMsg:
		if !m.scanning {
			m.scanning = true
			return m, tea.Batch(m.spinner.Tick, m.doScan())
		}
	}

	return m, nil
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m Model) View() string {
	if m.scanning {
		// In watch mode show the old tree behind the spinner overlay.
		if m.watch && m.output != "" {
			return m.renderRefreshOverlay()
		}
		return m.renderSpinner()
	}
	if m.watch {
		return m.output + m.renderWatchHint()
	}
	return m.output
}

// ── Rendering helpers ─────────────────────────────────────────────────────────

func (m Model) doScan() tea.Cmd {
	return func() tea.Msg {
		repos, err := scanner.Scan(m.scanRoot, m.opts)
		return scanDoneMsg{repos: repos, err: err}
	}
}

// renderSpinner shows the animated spinner while scanning from scratch.
func (m Model) renderSpinner() string {
	p, plain := theme.Load(m.cfg.Theme)
	path := collapseTilde(m.scanRoot)

	if plain {
		return fmt.Sprintf("Scanning %s...\n", path)
	}

	pathStyle := lipgloss.NewStyle().Foreground(p.HeaderPath).Bold(true)
	msgStyle := lipgloss.NewStyle().Foreground(p.HeaderFg).Faint(true)
	sep := lipgloss.NewStyle().Foreground(p.Separator).Render(strings.Repeat("─", m.width))

	return fmt.Sprintf("\n%s  %s %s\n%s\n",
		m.spinner.View(),
		msgStyle.Render("Scanning"),
		pathStyle.Render(path),
		sep,
	)
}

// renderRefreshOverlay renders the cached tree with a "refreshing" badge at top.
func (m Model) renderRefreshOverlay() string {
	p, plain := theme.Load(m.cfg.Theme)
	if plain {
		return m.output + "\nRefreshing...\n"
	}

	badge := lipgloss.NewStyle().
		Foreground(p.Ahead).
		Render(fmt.Sprintf("%s  Refreshing…", m.spinner.View()))

	sep := lipgloss.NewStyle().Foreground(p.Separator).Render(strings.Repeat("─", m.width))
	return badge + "\n" + sep + "\n" + m.output
}

// renderWatchHint renders the key-hint bar at the bottom in watch mode.
func (m Model) renderWatchHint() string {
	p, plain := theme.Load(m.cfg.Theme)

	age := ""
	if !m.lastUpdated.IsZero() {
		age = humanAge(time.Since(m.lastUpdated))
	}

	if plain {
		return fmt.Sprintf("\nUpdated %s  •  [r] refresh  [q] quit\n", age)
	}

	hintStyle := lipgloss.NewStyle().Foreground(p.FooterFg).Faint(true)
	keyStyle := lipgloss.NewStyle().Foreground(p.BranchFg)
	ageStyle := lipgloss.NewStyle().Foreground(p.CommitTime)

	hint := ageStyle.Render("Updated "+age) +
		hintStyle.Render("  •  ") +
		keyStyle.Render("r") + hintStyle.Render(" refresh  ") +
		keyStyle.Render("q") + hintStyle.Render(" quit") +
		hintStyle.Render(fmt.Sprintf("  •  refreshes every %s", watchInterval))

	sep := lipgloss.NewStyle().Foreground(p.Separator).Render(strings.Repeat("─", m.width))
	return sep + "\n" + hint + "\n"
}

// renderTree builds the full styled tree output.
func (m Model) renderTree() string {
	root := tree.Build(m.scanRoot, m.repos)
	renderer := tree.NewStyledRenderer(m.scanRoot, m.cfg.Theme)

	var sb strings.Builder
	if m.scanErr != nil {
		p, _ := theme.Load(m.cfg.Theme)
		warnStyle := lipgloss.NewStyle().Foreground(p.Error)
		sb.WriteString(warnStyle.Render("⚠  "+m.scanErr.Error()) + "\n")
	}
	renderer.Render(&sb, root)
	return sb.String()
}

// humanAge returns a concise relative string for durations < 1 minute,
// falling back to seconds for very fresh scans.
func humanAge(d time.Duration) string {
	switch {
	case d < time.Second:
		return "just now"
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	default:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
}

// collapseTilde replaces the home directory prefix with "~".
func collapseTilde(path string) string {
	home, _ := os.UserHomeDir()
	if home == "" {
		return path
	}
	if strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}
	return path
}

// ── Entry point ───────────────────────────────────────────────────────────────

// Run starts the Bubble Tea program and blocks until complete.
func Run(scanRoot string, opts scanner.Options, cfg *config.Config, watch bool) error {
	m := New(scanRoot, opts, cfg, watch)

	var p *tea.Program
	if watch {
		// Alt-screen keeps the tree contained and avoids scrollback pollution.
		p = tea.NewProgram(m, tea.WithAltScreen())
	} else {
		p = tea.NewProgram(m)
	}

	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	// In one-shot mode, print the rendered tree to scrollback after TUI exits.
	if !watch {
		if fm, ok := finalModel.(Model); ok {
			fmt.Print(fm.output)
		}
	}
	return nil
}
