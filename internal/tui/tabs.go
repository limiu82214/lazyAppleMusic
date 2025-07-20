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
	doc.WriteString(m.renderTabs())
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

func (m *tabTui) renderTabs() string {

	if len(m.Tabs) == 0 {
		return m.styles.docStyle.Render("No tabs available")
	}

	// 計算每個 tab 的寬度
	tabWidths := make([]int, len(m.Tabs))
	for i, tab := range m.Tabs {
		rendered := m.styles.inactiveTabStyle.Render(tab)
		tabWidths[i] = lipgloss.Width(rendered)
	}
	totalWidth := m.styles.width
	availableWidth := totalWidth

	// 完全不依賴 m.VisibleTabStart，動態計算 visibleTabStart
	visibleTabStart := 0
	end := 0
	currentWidth := 0

	// 先嘗試從頭開始塞滿
	for end < len(m.Tabs) && currentWidth+tabWidths[end] <= availableWidth {
		currentWidth += tabWidths[end]
		end++
	}
	// 如果 ActiveTab 不在視窗內，將視窗右移
	if m.ActiveTab < visibleTabStart {
		visibleTabStart = m.ActiveTab
		end = visibleTabStart
		currentWidth = 0
		for end < len(m.Tabs) && currentWidth+tabWidths[end] <= availableWidth {
			currentWidth += tabWidths[end]
			end++
		}
	} else if m.ActiveTab >= end {
		visibleTabStart = m.ActiveTab
		// 向左推到剛好能顯示滿寬度，且 activeTab 在視窗內
		for visibleTabStart > 0 {
			width := 0
			count := 0
			for i := visibleTabStart; i < len(m.Tabs) && width+tabWidths[i] <= availableWidth; i++ {
				width += tabWidths[i]
				count++
			}
			if m.ActiveTab >= visibleTabStart && m.ActiveTab < visibleTabStart+count {
				break
			}
			visibleTabStart--
		}
		end = visibleTabStart
		currentWidth = 0
		for end < len(m.Tabs) && currentWidth+tabWidths[end] <= availableWidth {
			currentWidth += tabWidths[end]
			end++
		}
	}

	pad := totalWidth
	for i := visibleTabStart; i < end; i++ {
		pad -= tabWidths[i]
	}

	visibleTabs := []string{}

	for i := visibleTabStart; i < end; i++ {
		style := m.styles.inactiveTabStyle
		if i == m.ActiveTab {
			style = m.styles.activeTabStyle
		}
		border, _, _, _, _ := style.GetBorder()

		isActive := i == m.ActiveTab
		isFirstVisible := i == visibleTabStart
		isLastVisible := i == end-1
		// isGloballyLast := i == len(m.Tabs)-1

		if isFirstVisible {
			if i == m.ActiveTab {

				border.BottomLeft = "│"
			} else {
				border.BottomLeft = "├"
			}
		}
		if isLastVisible {
			switch {
			case isLastVisible && isActive:
				border.BottomRight = "└"
			case pad > 0:
				border.BottomRight = "┴"
			case pad == 0:
				border.BottomRight = "┤"
			default:
				border.BottomRight = "?"
			}
		}
		style = style.Border(border)
		visibleTabs = append(visibleTabs, style.Render(m.Tabs[i]))
	}
	// 右收邊
	if pad > 0 {
		// 補線也要當成 box，丟進去
		visibleTabs = append(visibleTabs, "\n\n"+strings.Repeat("─", (pad+1))+"┐")
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, visibleTabs...)

}
