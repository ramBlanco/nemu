package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"nemu/internal/dns"
)

// View renders the input and the result for the hostname that was queried.
func (m Model) View() tea.View {
	var content strings.Builder
	fmt.Fprintln(&content, styles.title.Render("Nemu · DNS lookup"))
	fmt.Fprintln(&content)
	fmt.Fprintf(&content, "Hostname %s\n", m.input.View())
	fmt.Fprintln(&content, "Enter: consultar  ·  Ctrl+C/Esc: salir")

	switch {
	case m.loading:
		fmt.Fprintf(&content, "\nGuardando y consultando %s…\n", strconv.QuoteToGraphic(m.queriedHost))
	case m.hasResult:
		fmt.Fprintf(&content, "\nResultado para %s\n", strconv.QuoteToGraphic(m.queriedHost))
		fmt.Fprintf(&content, "Estado: %s\n", styles.forOutcome(m.outcome.Status).Render(outcomeLabel(m.outcome.Status)))
		switch m.outcome.Status {
		case dns.OutcomeResolved:
			if len(m.outcome.Addresses) > 0 {
				fmt.Fprintf(&content, "Direcciones: %s\n", strings.Join(m.outcome.Addresses, ", "))
			}
			if m.outcome.Err != nil {
				fmt.Fprintf(&content, "Aviso de respuesta parcial: %s\n", strconv.QuoteToGraphic(m.outcome.Err.Error()))
			}
		case dns.OutcomeInconclusive:
			if m.outcome.Err != nil {
				fmt.Fprintf(&content, "Detalle: %s\n", strconv.QuoteToGraphic(m.outcome.Err.Error()))
			}
		}
	}

	fmt.Fprintln(&content)
	fmt.Fprintln(&content, styles.title.Render("Hostnames guardados"))
	if !m.storeLoaded {
		fmt.Fprintln(&content, styles.muted.Render("Cargando hostnames…"))
	} else {
		if m.storeErr != nil {
			fmt.Fprintf(&content, "%s %s\n", styles.storageError.Render("No se pudo guardar/cargar:"), strconv.QuoteToGraphic(m.storeErr.Error()))
		}
		if len(m.hostnames) == 0 {
			fmt.Fprintln(&content, styles.muted.Render("(ninguno)"))
		} else {
			for _, hostname := range m.hostnames {
				fmt.Fprintf(&content, "• %s\n", strconv.QuoteToGraphic(hostname))
			}
		}
	}

	view := tea.NewView(content.String())
	view.AltScreen = true
	return view
}
