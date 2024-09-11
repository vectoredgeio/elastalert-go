package models

import "encoding/json"

type Detection struct {
	ID                     int             `json:"ID"`
	CreatedAt              string          `json:"CreatedAt"`
	UpdatedAt              string          `json:"UpdatedAt"`
	DeletedAt              *string         `json:"DeletedAt"` // Nullable
	TenantID               int             `json:"tenant_id"`
	CampaignID             int             `json:"campaign_id"`
	DetectionName          string          `json:"detection_name"`
	Service                string          `json:"service"`
	DetectionAccessibility string          `json:"detection_accessibility"`
	QueryString            json.RawMessage `json:"query_string"` // String containing JSON
	AlgorithmType          string          `json:"algorithm_type"`
	DetectionQueryString   string          `json:"detection_query_string"`
}

type QueryString struct {
	NodeSelection      NodeSelection       `json:"node_selection"`
	EdgeSelectionSteps []EdgeSelectionStep `json:"edge_selection_steps"`
}
type NodeSelection struct {
	Node       int64       `json:"node"`
	NodeTypes  []string    `json:"node_types"`
	Conditions []Condition `json:"conditions"`
}

type EdgeSelectionStep struct {
	Edges      []string    `json:"edges"`
	Conditions []Condition `json:"conditions"`
}
type Condition struct {
	NodeType string `json:"node_type"`
	Property string `json:"property"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}
