package bbt_client

import (
	tea "charm.land/bubbletea/v2"
	"github.com/tobyd02/golang-mmo/pkg/client"
)

func BBTReadGameWorldDiff(c *client.GClient) tea.Cmd {
	return func() tea.Msg {
		worldDiff, err := c.ReadGameWorldDiff()

		if err != nil {
			return ConnectionErrorMsg{Err: err}
		}

		return worldDiff
	}
}

func BBTSendMoveMessage(c *client.GClient, dx, dy int) tea.Cmd {
	return func() tea.Msg {
		err := c.SendMoveMessage(dx, dy)
		if err != nil {
			return ConnectionErrorMsg{Err: err}
		}

		return nil
	}
}

func BBTSendInteractMessage(c *client.GClient, interactableInstanceID string) tea.Cmd {
	return func() tea.Msg {
		err := c.SendInteractMessage(interactableInstanceID)
		if err != nil {
			return ConnectionErrorMsg{Err: err}
		}

		return nil
	}
}

func BBTSendAttackNpcMessage(c *client.GClient, npcInstanceID string) tea.Cmd {
	return func() tea.Msg {
		err := c.SendAttackNpcMessage(npcInstanceID)
		if err != nil {
			return ConnectionErrorMsg{Err: err}
		}

		return nil
	}
}
