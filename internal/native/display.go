package native

import (
	"slices"
	"time"
)

func (n *Native) setUIVars() {
	if n.appVersion != nil {
		uiSetVar("app_version", n.appVersion.String())
	} else {
		uiSetVar("app_version", "unknown")
	}

	if n.systemVersion != nil {
		uiSetVar("system_version", n.systemVersion.String())
	} else {
		uiSetVar("system_version", "unknown")
	}
}

func (n *Native) initUI() {
	// Skip UI initialization for headless operation
	n.l.Info().Msg("UI initialization skipped - headless operation")
	// Still set UI vars for compatibility, but with nil checks
	n.setUIVars()
}

func (n *Native) tickUI() {
	// Skip UI ticking for headless operation
	n.l.Info().Msg("UI ticking skipped - headless operation")
	// Just sleep to prevent busy loop
	for {
		time.Sleep(100 * time.Millisecond)
	}
}

// GetLVGLVersion returns the LVGL version
func (n *Native) GetLVGLVersion() (string, error) {
	return uiGetLVGLVersion(), nil
}

// UIObjHide hides the object
func (n *Native) UIObjHide(objName string) (bool, error) {
	return uiObjHide(objName)
}

// UIObjShow shows the object
func (n *Native) UIObjShow(objName string) (bool, error) {
	return uiObjShow(objName)
}

// UISetVar sets the variable
func (n *Native) UISetVar(name string, value string) {
	uiSetVar(name, value)
}

// UIGetVar gets the variable
func (n *Native) UIGetVar(name string) string {
	return uiGetVar(name)
}

// UIObjAddState adds the state to the object
func (n *Native) UIObjAddState(objName string, state string) (bool, error) {
	return uiObjAddState(objName, state)
}

// UIObjClearState clears the state from the object
func (n *Native) UIObjClearState(objName string, state string) (bool, error) {
	return uiObjClearState(objName, state)
}

// UIObjAddFlag adds the flag to the object
func (n *Native) UIObjAddFlag(objName string, flag string) (bool, error) {
	return uiObjAddFlag(objName, flag)
}

// UIObjClearFlag clears the flag from the object
func (n *Native) UIObjClearFlag(objName string, flag string) (bool, error) {
	return uiObjClearFlag(objName, flag)
}

// UIObjSetOpacity sets the opacity of the object
func (n *Native) UIObjSetOpacity(objName string, opacity int) (bool, error) {
	return uiObjSetOpacity(objName, opacity)
}

// UIObjFadeIn fades in the object
func (n *Native) UIObjFadeIn(objName string, duration uint32) (bool, error) {
	return uiObjFadeIn(objName, duration)
}

// UIObjFadeOut fades out the object
func (n *Native) UIObjFadeOut(objName string, duration uint32) (bool, error) {
	return uiObjFadeOut(objName, duration)
}

// UIObjSetLabelText sets the text of the object
func (n *Native) UIObjSetLabelText(objName string, text string) (bool, error) {
	return uiLabelSetText(objName, text)
}

// UIObjSetImageSrc sets the image of the object
func (n *Native) UIObjSetImageSrc(objName string, image string) (bool, error) {
	return uiImgSetSrc(objName, image)
}

// DisplaySetRotation sets the rotation of the display
func (n *Native) DisplaySetRotation(rotation uint16) (bool, error) {
	return uiDispSetRotation(rotation)
}

// UpdateLabelIfChanged updates the label if the text has changed
func (n *Native) UpdateLabelIfChanged(objName string, newText string) {
	l := n.lD.Trace().Str("obj", objName).Str("text", newText)

	changed, err := n.UIObjSetLabelText(objName, newText)
	if err != nil {
		n.lD.Warn().Str("obj", objName).Str("text", newText).Err(err).Msg("failed to update label")
		return
	}

	if changed {
		l.Msg("label changed")
	} else {
		l.Msg("label not changed")
	}
}

// UpdateLabelAndChangeVisibility updates the label and changes the visibility of the object
func (n *Native) UpdateLabelAndChangeVisibility(objName string, newText string) {
	n.UpdateLabelIfChanged(objName, newText)

	containerName := objName + "_container"
	if newText == "" {
		_, _ = n.UIObjHide(objName)
		_, _ = n.UIObjHide(containerName)
	} else {
		_, _ = n.UIObjShow(objName)
		_, _ = n.UIObjShow(containerName)
	}
}

// SwitchToScreenIf switches to the screen if the screen name is different from the current screen and the screen name is in the shouldSwitch list
func (n *Native) SwitchToScreenIf(screenName string, shouldSwitch []string) {
	currentScreen := uiGetCurrentScreen()
	if currentScreen == screenName {
		return
	}
	if len(shouldSwitch) > 0 && !slices.Contains(shouldSwitch, currentScreen) {
		n.lD.Trace().Str("from", currentScreen).Str("to", screenName).Msg("skipping screen switch")
		return
	}
	n.lD.Info().Str("from", currentScreen).Str("to", screenName).Msg("switching screen")
	uiSwitchToScreen(screenName)
}

// SwitchToScreenIfDifferent switches to the screen if the screen name is different from the current screen
func (n *Native) SwitchToScreenIfDifferent(screenName string) {
	n.SwitchToScreenIf(screenName, []string{})
}
