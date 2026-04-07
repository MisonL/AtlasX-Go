package tabs

import "fmt"

func applyDeviceEmulation(websocketURL string, preset DevicePreset) error {
	if _, err := runPageCommand(websocketURL, cdpCommandRequest{
		ID:     10,
		Method: "Emulation.setDeviceMetricsOverride",
		Params: map[string]any{
			"width":             preset.Width,
			"height":            preset.Height,
			"deviceScaleFactor": preset.DeviceScaleFactor,
			"mobile":            preset.Mobile,
		},
	}); err != nil {
		return err
	}
	_, err := runPageCommand(websocketURL, cdpCommandRequest{
		ID:     11,
		Method: "Emulation.setTouchEmulationEnabled",
		Params: map[string]any{
			"enabled":        preset.Touch,
			"maxTouchPoints": maxTouchPointsForPreset(preset),
		},
	})
	return err
}

func clearDeviceEmulation(websocketURL string) error {
	if _, err := runPageCommand(websocketURL, cdpCommandRequest{
		ID:     12,
		Method: "Emulation.clearDeviceMetricsOverride",
	}); err != nil {
		return err
	}
	_, err := runPageCommand(websocketURL, cdpCommandRequest{
		ID:     13,
		Method: "Emulation.setTouchEmulationEnabled",
		Params: map[string]any{
			"enabled":        false,
			"maxTouchPoints": 0,
		},
	})
	return err
}

func maxTouchPointsForPreset(preset DevicePreset) int {
	if preset.Touch {
		return 5
	}
	return 0
}

func formatDeviceViewport(preset DevicePreset) string {
	return fmt.Sprintf("%dx%d@%g", preset.Width, preset.Height, preset.DeviceScaleFactor)
}
