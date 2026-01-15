package kvm

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/xkvm/kvm/internal/confparser"
	"github.com/xkvm/kvm/internal/logging"
	"github.com/xkvm/kvm/internal/network/types"
	"github.com/xkvm/kvm/internal/usbgadget"
)

const (
	DefaultAPIURL = "https://api.xkvm.com"
)

type WakeOnLanDevice struct {
	Name       string `json:"name"`
	MacAddress string `json:"macAddress"`
}

// Constants for keyboard macro limits
const (
	MaxMacrosPerDevice = 25
	MaxStepsPerMacro   = 10
	MaxKeysPerStep     = 10
	MinStepDelay       = 50
	MaxStepDelay       = 2000
)

type KeyboardMacroStep struct {
	Keys      []string `json:"keys"`
	Modifiers []string `json:"modifiers"`
	Delay     int      `json:"delay"`
}

func (s *KeyboardMacroStep) Validate() error {
	if len(s.Keys) > MaxKeysPerStep {
		return fmt.Errorf("too many keys in step (max %d)", MaxKeysPerStep)
	}

	if s.Delay < MinStepDelay {
		s.Delay = MinStepDelay
	} else if s.Delay > MaxStepDelay {
		s.Delay = MaxStepDelay
	}

	return nil
}

type KeyboardMacro struct {
	ID        string              `json:"id"`
	Name      string              `json:"name"`
	Steps     []KeyboardMacroStep `json:"steps"`
	SortOrder int                 `json:"sortOrder,omitempty"`
}

func (m *KeyboardMacro) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("macro name cannot be empty")
	}

	if len(m.Steps) == 0 {
		return fmt.Errorf("macro must have at least one step")
	}

	if len(m.Steps) > MaxStepsPerMacro {
		return fmt.Errorf("too many steps in macro (max %d)", MaxStepsPerMacro)
	}

	for i := range m.Steps {
		if err := m.Steps[i].Validate(); err != nil {
			return fmt.Errorf("invalid step %d: %w", i+1, err)
		}
	}

	return nil
}

type Config struct {
	CloudURL           string               `json:"cloud_url"`
	JigglerEnabled     bool                 `json:"jiggler_enabled"`
	JigglerConfig      *JigglerConfig       `json:"jiggler_config"`
	AutoUpdateEnabled  bool                 `json:"auto_update_enabled"`
	IncludePreRelease  bool                 `json:"include_pre_release"`
	HashedPassword     string               `json:"hashed_password"`
	LocalAuthToken     string               `json:"local_auth_token"`
	LocalAuthMode      string               `json:"localAuthMode"` //TODO: fix it with migration
	LocalLoopbackOnly  bool                 `json:"local_loopback_only"`
	WakeOnLanDevices   []WakeOnLanDevice    `json:"wake_on_lan_devices"`
	KeyboardMacros     []KeyboardMacro      `json:"keyboard_macros"`
	KeyboardLayout     string               `json:"keyboard_layout"`
	EdidString         string               `json:"hdmi_edid_string"`
	ActiveExtension    string               `json:"active_extension"`
	TLSMode            string               `json:"tls_mode"` // options: "self-signed", "user-defined", ""
	UsbConfig          *usbgadget.Config    `json:"usb_config"`
	UsbDevices         *usbgadget.Devices   `json:"usb_devices"`
	NetworkConfig      *types.NetworkConfig `json:"network_config,omitempty"` // Deprecated: Network config is read-only on full Linux systems. This field is ignored during save/load.
	DefaultLogLevel    string               `json:"default_log_level"`
	VideoSleepAfterSec int                  `json:"video_sleep_after_sec"`
	VideoQualityFactor float64              `json:"video_quality_factor"`
	NativeMaxRestart   uint                 `json:"native_max_restart_attempts"`
}

const configPath = "/userdata/kvm_config.json"

// it's a temporary solution to avoid sharing the same pointer
// we should migrate to a proper config solution in the future
var (
	defaultJigglerConfig = JigglerConfig{
		InactivityLimitSeconds: 60,
		JitterPercentage:       25,
		ScheduleCronTab:        "0 * * * * *",
		Timezone:               "UTC",
	}
	defaultUsbConfig = usbgadget.Config{
		VendorId:     "0x1d6b", //The Linux Foundation
		ProductId:    "0x0104", //Multifunction Composite Gadget
		SerialNumber: "",
		Manufacturer: "XKVM",
		Product:      "USB Emulation Device",
	}
	defaultUsbDevices = usbgadget.Devices{
		AbsoluteMouse: true,
		RelativeMouse: true,
		Keyboard:      true,
		MassStorage:   true,
	}
)

func getDefaultConfig() Config {
	return Config{
		CloudURL:          DefaultAPIURL,
		AutoUpdateEnabled: true, // Set a default value
		ActiveExtension:   "",
		KeyboardMacros:    []KeyboardMacro{},
		KeyboardLayout:    "en-US",
		JigglerEnabled:    false,
		// This is the "Standard" jiggler option in the UI
		JigglerConfig: func() *JigglerConfig { c := defaultJigglerConfig; return &c }(),
		TLSMode:       "",
		UsbConfig:     func() *usbgadget.Config { c := defaultUsbConfig; return &c }(),
		UsbDevices:    func() *usbgadget.Devices { c := defaultUsbDevices; return &c }(),
		NetworkConfig: func() *types.NetworkConfig {
			c := &types.NetworkConfig{}
			_ = confparser.SetDefaultsAndValidate(c)
			return c
		}(),
		DefaultLogLevel:    "INFO",
		VideoQualityFactor: 5000.0,
	}
}

var (
	config     *Config
	configLock = &sync.Mutex{}
)

var (
	configSuccess = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "xkvm_config_last_reload_successful",
			Help: "The last configuration load succeeded",
		},
	)
	configSuccessTime = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "xkvm_config_last_reload_success_timestamp_seconds",
			Help: "Timestamp of last successful config load",
		},
	)
)

func LoadConfig() {
	configLock.Lock()
	defer configLock.Unlock()

	if config != nil {
		logger.Debug().Msg("config already loaded, skipping")
		return
	}

	// load the default config
	defaultConfig := getDefaultConfig()
	config = &defaultConfig

	file, err := os.Open(configPath)
	if err != nil {
		logger.Debug().Msg("default config file doesn't exist, using default")
		configSuccess.Set(1.0)
		configSuccessTime.SetToCurrentTime()
		return
	}
	defer file.Close()

	// load and merge the default config with the user config
	loadedConfig := defaultConfig
	if err := json.NewDecoder(file).Decode(&loadedConfig); err != nil {
		logger.Warn().Err(err).Msg("config file JSON parsing failed")
		configSuccess.Set(0.0)
		return
	}

	// merge the user config with the default config
	if loadedConfig.UsbConfig == nil {
		loadedConfig.UsbConfig = getDefaultConfig().UsbConfig
	}

	if loadedConfig.UsbDevices == nil {
		loadedConfig.UsbDevices = getDefaultConfig().UsbDevices
	}

	// Network config migration: Always use default NetworkConfig in memory
	// The persisted network_config is ignored as network is now read-only
	loadedConfig.NetworkConfig = getDefaultConfig().NetworkConfig
	logger.Debug().Msg("network config loaded from defaults (persisted config ignored)")

	if loadedConfig.JigglerConfig == nil {
		loadedConfig.JigglerConfig = getDefaultConfig().JigglerConfig
	}

	// fixup old keyboard layout value
	if loadedConfig.KeyboardLayout == "en_US" {
		loadedConfig.KeyboardLayout = "en-US"
	}

	// migrate old quality factor values (0.1, 0.5, 1.0) to new bitrate format
	if loadedConfig.VideoQualityFactor < 10 {
		loadedConfig.VideoQualityFactor = 5000.0
	}

	config = &loadedConfig

	logging.GetRootLogger().UpdateLogLevel(config.DefaultLogLevel)

	configSuccess.Set(1.0)
	configSuccessTime.SetToCurrentTime()

	logger.Info().Str("path", configPath).Msg("config loaded")
}

func SaveConfig() error {
	return saveConfig(configPath)
}

func SaveBackupConfig() error {
	return saveConfig(configPath + ".bak")
}

func saveConfig(path string) error {
	configLock.Lock()
	defer configLock.Unlock()

	logger.Trace().Str("path", path).Msg("Saving config")

	// fixup old keyboard layout value
	if config.KeyboardLayout == "en_US" {
		config.KeyboardLayout = "en-US"
	}

	// Create a copy of config for saving, excluding NetworkConfig
	// Network config is read-only and should not be persisted
	configToSave := *config
	configToSave.NetworkConfig = nil

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(configToSave); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to wite config: %w", err)
	}

	logger.Info().Str("path", path).Msg("config saved (network_config excluded)")
	return nil
}

func ensureConfigLoaded() {
	if config == nil {
		LoadConfig()
	}
}
