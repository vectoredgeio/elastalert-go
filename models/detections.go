package models

type Detection struct {
	ID                     int     `json:"ID"`
	CreatedAt              string  `json:"CreatedAt"`
	UpdatedAt              string  `json:"UpdatedAt"`
	DeletedAt              *string `json:"DeletedAt"`
	TenantID               int     `json:"tenant_id"`
	CampaignID             int     `json:"campaign_id"`
	DetectionName          string  `json:"detection_name"`
	Service                string  `json:"service"`
	DetectionAccessibility string  `json:"detection_accessibility"`
	QueryString            string  `json:"query_string"`
	AlgorithmType          string  `json:"algorithm_type"`
	DetectionQueryString   string  `json:"detection_query_string"`
}

type DetectionScore struct {
	ID                     int    `json:"ID"`
	TenantID               int    `json:"tenant_id"`
	CampaignID             int    `json:"campaign_id"`
	DetectionName          string `json:"detection_name"`
	Service                string `json:"service"`
	DetectionAccessibility string `json:"detection_accessibility"`
	QueryString            string `json:"query_string"`
	AlgorithmType          string `json:"algorithm_type"`
	DetectionQueryString   string `json:"detection_query_string"`
	AES                    int    `json:"aes"`
}
type AESResponse struct {
	Status  bool     `json:"status"`
	Data    int      `json:"data"` // Assuming AES score is an integer
	Message string   `json:"message"`
	Error   []string `json:"error"`
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

func (detection Detection) TransformIntoDetectionScore(calcDetection Detection, aes int) DetectionScore {
	return DetectionScore{
		ID:                     detection.ID,
		TenantID:               detection.TenantID,
		CampaignID:             detection.CampaignID,
		DetectionName:          detection.DetectionName,
		Service:                detection.Service,
		DetectionAccessibility: detection.DetectionAccessibility,
		QueryString:            detection.QueryString,
		AlgorithmType:          detection.AlgorithmType,
		DetectionQueryString:   detection.DetectionQueryString,
		AES:                    aes,
	}
}
