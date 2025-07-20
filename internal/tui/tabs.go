package tui

import (
	"io"
	"limiu82214/lazyAppleMusic/internal/constant"
	"limiu82214/lazyAppleMusic/internal/model"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davecgh/go-spew/spew"
)

var tabDebug = true

type TabTui interface {
	tea.Model

	NextPage() TabTui
	PrevPage() TabTui
	SetHeight(height int) TabTui
	SetWidth(width int) TabTui
}
type tabTui struct {
	dump io.Writer
	model.TabTuiData

	styles tabStyles
}

type tabStyles struct {
	inactiveTabBorder lipgloss.Border
	activeTabBorder   lipgloss.Border
	docStyle          lipgloss.Style
	highlightColor    lipgloss.AdaptiveColor
	inactiveTabStyle  lipgloss.Style
	activeTabStyle    lipgloss.Style
	windowStyle       lipgloss.Style
	width             int
	height            int
}

func newTabTui(dump io.Writer, tabs []string, tabContent []tea.Model, activeTab int) TabTui {
	if len(tabs) != len(tabContent) {
		panic("tabs and tabContent must have the same length")
	}
	obj := &tabTui{
		dump: dump,
		TabTuiData: model.TabTuiData{
			Tabs:       tabs,
			TabContent: tabContent,
			ActiveTab:  activeTab,
		},
	}
	ts := tabStyles{
		inactiveTabBorder: obj.tabBorderWithBottom("┴", "─", "┴"),
		activeTabBorder:   obj.tabBorderWithBottom("┘", " ", "└"),
		docStyle:          lipgloss.NewStyle().Padding(0, 0, 0, 0),
		highlightColor:    lipgloss.AdaptiveColor{Light: "#fd4b60ff", Dark: "#ffffffff"},
	}
	ts.inactiveTabStyle = lipgloss.NewStyle().Border(ts.inactiveTabBorder, true).BorderForeground(ts.highlightColor).Padding(0, 1)
	ts.activeTabStyle = ts.inactiveTabStyle.Border(ts.activeTabBorder, true)
	ts.windowStyle = lipgloss.NewStyle().BorderForeground(ts.highlightColor).Padding(0, 0).Align(lipgloss.Left).Border(lipgloss.RoundedBorder()).UnsetBorderTop()

	obj.styles = ts
	if !tabDebug {
		obj.dump = io.Discard
	}
	return obj
}

// ======= MAIN

func (m *tabTui) Init() tea.Cmd {
	return nil
}

func (m *tabTui) View() string {
	if len(m.Tabs) == 0 {
		return m.styles.docStyle.Render("No tabs available")
	}

	doc := strings.Builder{}
	var renderedTabs []string
	for i, tab := range m.Tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.Tabs)-1, i == m.ActiveTab
		if i == m.ActiveTab {
			style = m.styles.activeTabStyle
		} else {
			style = m.styles.inactiveTabStyle
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(tab))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")
	spew.Fprintln(m.dump, "width", m.styles.width, m.styles.windowStyle.GetHorizontalFrameSize(), m.styles.windowStyle.GetHorizontalBorderSize())
	window := m.styles.windowStyle.Width(m.styles.width).
		Height(m.styles.height - lipgloss.Height(doc.String()))

	if ml, ok := m.TabContent[m.ActiveTab].(CurrentPlaylistTui); ok {
		ml.SetHeight(window.GetHeight() - m.styles.windowStyle.GetVerticalBorderSize()).
			SetWidth(window.GetWidth() - m.styles.windowStyle.GetHorizontalFrameSize())
		m.TabContent[m.ActiveTab] = ml
	}

	doc.WriteString(window.Render(m.TabContent[m.ActiveTab].View()))

	return m.styles.docStyle.Render(doc.String())

}

func (m *tabTui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := []tea.Cmd{}
	switch msg := msg.(type) {
	case constant.ShouldUpdateTabs:
		m.TabTuiData = model.TabTuiData(msg)

	case tea.WindowSizeMsg:
		m.styles.width = msg.Width
		m.styles.height = msg.Height
		spew.Fprintln(m.dump, "QQQQ22222222", msg.Height)
		var cmd tea.Cmd
		for i := range m.TabContent {
			m.TabContent[i], cmd = m.TabContent[i].Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	default:
		for i := range m.TabContent {
			c, cmd := m.TabContent[i].Update(msg)
			m.TabContent[i] = c
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

// other

func (m *tabTui) tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

func (m *tabTui) SetHeight(height int) TabTui {
	m.styles.height = height
	return m
}
func (m *tabTui) SetWidth(width int) TabTui {
	m.styles.width = width
	return m
}
func (m *tabTui) NextPage() TabTui {
	if m.ActiveTab < len(m.TabContent)-1 {
		m.ActiveTab++
	}
	return m
}
func (m *tabTui) PrevPage() TabTui {
	if m.ActiveTab > 0 {
		m.ActiveTab--
	}
	return m
}
