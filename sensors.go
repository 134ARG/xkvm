package kvm

import (
	"context"
	"encoding/json"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SensorFeature struct {
	Path  string  `json:"path"`
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

type SensorChip struct {
	Chip     string          `json:"chip"`
	Adapter  string          `json:"adapter"`
	Features []SensorFeature `json:"features"`
}

type SensorConfig struct {
	EnvChip     string `json:"envChip"`
	TempFeature string `json:"tempFeature"`
	HumFeature  string `json:"humFeature"`
	SocChip     string `json:"socChip"`
	SocFeature  string `json:"socFeature"`
}

type EnvironmentMetrics struct {
	CaseTemperatureC   *float64 `json:"caseTemperatureC"`
	CaseHumidityPct    *float64 `json:"caseHumidityPercent"`
	SocTemperatureC    *float64 `json:"socTemperatureC"`
	UpdatedAtUnixMilli int64    `json:"updatedAt"`
}

var sensorCache = struct {
	sync.Mutex
	at    time.Time
	chips []SensorChip
	err   error
}{}

func rpcGetSensorChips() ([]SensorChip, error) {
	return readSensorChips()
}

func rpcGetSensorConfig() (SensorConfig, error) {
	return SensorConfig{
		EnvChip:     config.SensorEnvChip,
		TempFeature: config.SensorTempFeature,
		HumFeature:  config.SensorHumFeature,
		SocChip:     config.SensorSocChip,
		SocFeature:  config.SensorSocFeature,
	}, nil
}

func rpcSetSensorConfig(sensorConfig SensorConfig) error {
	config.SensorEnvChip = sensorConfig.EnvChip
	config.SensorTempFeature = sensorConfig.TempFeature
	config.SensorHumFeature = sensorConfig.HumFeature
	config.SensorSocChip = sensorConfig.SocChip
	config.SensorSocFeature = sensorConfig.SocFeature
	return SaveConfig()
}

func rpcGetEnvironmentMetrics() (EnvironmentMetrics, error) {
	chips, err := readSensorChips()
	if err != nil {
		return EnvironmentMetrics{UpdatedAtUnixMilli: time.Now().UnixMilli()}, nil
	}

	metrics := EnvironmentMetrics{
		CaseTemperatureC:   lookupSensorValue(chips, config.SensorEnvChip, config.SensorTempFeature),
		CaseHumidityPct:    lookupSensorValue(chips, config.SensorEnvChip, config.SensorHumFeature),
		SocTemperatureC:    lookupSensorValue(chips, config.SensorSocChip, config.SensorSocFeature),
		UpdatedAtUnixMilli: time.Now().UnixMilli(),
	}
	return metrics, nil
}

func readSensorChips() ([]SensorChip, error) {
	sensorCache.Lock()
	if time.Since(sensorCache.at) < 2*time.Second {
		chips, err := sensorCache.chips, sensorCache.err
		sensorCache.Unlock()
		return chips, err
	}
	sensorCache.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, "sensors", "-j").Output()
	var chips []SensorChip
	if err == nil {
		chips, err = parseSensorsJSON(out)
	}

	sensorCache.Lock()
	sensorCache.at = time.Now()
	sensorCache.chips = chips
	sensorCache.err = err
	sensorCache.Unlock()

	return chips, err
}

func parseSensorsJSON(data []byte) ([]SensorChip, error) {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	chips := make([]SensorChip, 0, len(raw))
	for chipName, chipRaw := range raw {
		chipMap, ok := chipRaw.(map[string]any)
		if !ok {
			continue
		}

		chip := SensorChip{Chip: chipName}
		if adapter, ok := chipMap["Adapter"].(string); ok {
			chip.Adapter = adapter
		}

		for groupName, groupRaw := range chipMap {
			groupMap, ok := groupRaw.(map[string]any)
			if !ok {
				continue
			}

			for inputName, valueRaw := range groupMap {
				value, ok := numberFromAny(valueRaw)
				if !ok {
					continue
				}
				path := groupName + "." + inputName
				chip.Features = append(chip.Features, SensorFeature{
					Path:  path,
					Label: strings.ReplaceAll(path, "_", " "),
					Value: value,
				})
			}
		}

		sort.Slice(chip.Features, func(i, j int) bool {
			return chip.Features[i].Path < chip.Features[j].Path
		})
		chips = append(chips, chip)
	}

	sort.Slice(chips, func(i, j int) bool {
		return chips[i].Chip < chips[j].Chip
	})
	return chips, nil
}

func lookupSensorValue(chips []SensorChip, chipName string, featurePath string) *float64 {
	if chipName == "" || featurePath == "" {
		return nil
	}
	for _, chip := range chips {
		if chip.Chip != chipName {
			continue
		}
		for _, feature := range chip.Features {
			if feature.Path == featurePath {
				value := feature.Value
				return &value
			}
		}
	}
	return nil
}

func numberFromAny(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case string:
		n, err := strconv.ParseFloat(v, 64)
		return n, err == nil
	case json.Number:
		n, err := v.Float64()
		return n, err == nil
	default:
		return 0, false
	}
}
