package api

const (
	LiveStatus     = "OK"
	ReadyStatus    = "READY"
	NotReadyStatus = "NOT READY"
)

type HealthResponse struct {
	Status string `json:"status"`
}

func NewHealthResponse(status string) *HealthResponse {
	return &HealthResponse{
		Status: status,
	}
}
