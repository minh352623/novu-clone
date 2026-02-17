package dto

import (
	"time"

	"CONVERDA/internal/messaging/domain/model/entity"

	"github.com/google/uuid"
)

type DashboardStatsRequest struct {
	EnvironmentID uuid.UUID `form:"environment_id" binding:"required"`
	From          time.Time `form:"from"`
	To            time.Time `form:"to"`
	SLAThreshold  int       `form:"sla_threshold"` // in seconds
}

type TeamDashboardResponse struct {
	Stats *entity.TeamStats `json:"stats"`
}

type PartnerDashboardResponse struct {
	Stats *entity.TeamStats `json:"stats"` // Partner dashboard uses TeamStats for now
}

type AgentDashboardResponse struct {
	Stats *entity.AgentStats `json:"stats"`
}

type ActivityPoint struct {
	Time  time.Time `json:"time"`
	Value int       `json:"value"`
}

type PersonalDashboardResponse struct {
	Stats            *entity.AgentStats   `json:"stats"`
	ActivityTimeline []ActivityPoint      `json:"activity_timeline"`
	TeamComparison   *entity.TeamStats    `json:"team_comparison"`
	TopPerformers    []*entity.AgentStats `json:"top_performers,omitempty"`
}
