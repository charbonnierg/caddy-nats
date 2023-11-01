package service

import "encoding/json"

type Logs struct {
	Level         string            `json:"level,omitempty"`
	InitialFields map[string]string `json:"initial_fields,omitempty"`
}

type Metrics struct {
	Level   string `json:"level,omitempty"`
	Address string `json:"address,omitempty"`
}

type Telemetry struct {
	Logs    Logs    `json:"logs,omitempty"`
	Metrics Metrics `json:"metrics,omitempty"`
}

type Pipeline struct {
	Receivers  []string `json:"receivers,omitempty"`
	Processors []string `json:"processors,omitempty"`
	Exporters  []string `json:"exporters,omitempty"`
}

type Pipelines struct {
	Traces  Pipeline `json:"traces,omitempty"`
	Metrics Pipeline `json:"metrics,omitempty"`
	Logs    Pipeline `json:"logs,omitempty"`
}

type Service struct {
	Extensions []string  `json:"extensions,omitempty"`
	Pipelines  Pipelines `json:"pipelines,omitempty"`
	Telemetry  Telemetry `json:"telemetry,omitempty"`
}

func (s *Service) Values() (map[string]any, error) {
	raw, err := json.Marshal(&s)
	if err != nil {
		return nil, err
	}
	var values map[string]any
	err = json.Unmarshal(raw, &values)
	if err != nil {
		return nil, err
	}
	return values, nil
}
