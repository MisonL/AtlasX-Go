package tabs

import "errors"

type DevicePreset struct {
	Name              string
	Width             int
	Height            int
	DeviceScaleFactor float64
	Mobile            bool
	Touch             bool
}

type DeviceEmulationResult struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	URL      string `json:"url"`
	Preset   string `json:"preset"`
	Viewport string `json:"viewport"`
	Mobile   bool   `json:"mobile"`
	Touch    bool   `json:"touch"`
}

var devicePresets = map[string]DevicePreset{
	"iphone-13": {
		Name:              "iphone-13",
		Width:             390,
		Height:            844,
		DeviceScaleFactor: 3,
		Mobile:            true,
		Touch:             true,
	},
	"ipad-mini": {
		Name:              "ipad-mini",
		Width:             768,
		Height:            1024,
		DeviceScaleFactor: 2,
		Mobile:            true,
		Touch:             true,
	},
	"desktop-wide": {
		Name:              "desktop-wide",
		Width:             1440,
		Height:            900,
		DeviceScaleFactor: 1,
		Mobile:            false,
		Touch:             false,
	},
}

func lookupDevicePreset(name string) (DevicePreset, bool) {
	preset, ok := devicePresets[name]
	return preset, ok
}

func (c Client) EmulateDevice(targetID string, presetName string) (DeviceEmulationResult, error) {
	if presetName != "off" {
		if _, ok := lookupDevicePreset(presetName); !ok {
			return DeviceEmulationResult{}, errors.New("unknown device preset " + `"` + presetName + `"`)
		}
	}

	targets, err := c.List()
	if err != nil {
		return DeviceEmulationResult{}, err
	}

	target, err := findPageTarget(targets, targetID)
	if err != nil {
		return DeviceEmulationResult{}, err
	}
	if target.WebSocketDebuggerURL == "" {
		return DeviceEmulationResult{}, errors.New("target does not expose a websocket debugger url")
	}

	if presetName == "off" {
		if err := clearDeviceEmulation(target.WebSocketDebuggerURL); err != nil {
			return DeviceEmulationResult{}, err
		}
		return DeviceEmulationResult{
			ID:       target.ID,
			Title:    target.Title,
			URL:      target.URL,
			Preset:   "off",
			Viewport: "off",
			Mobile:   false,
			Touch:    false,
		}, nil
	}

	preset, _ := lookupDevicePreset(presetName)
	if err := applyDeviceEmulation(target.WebSocketDebuggerURL, preset); err != nil {
		return DeviceEmulationResult{}, err
	}
	return DeviceEmulationResult{
		ID:       target.ID,
		Title:    target.Title,
		URL:      target.URL,
		Preset:   preset.Name,
		Viewport: formatDeviceViewport(preset),
		Mobile:   preset.Mobile,
		Touch:    preset.Touch,
	}, nil
}
