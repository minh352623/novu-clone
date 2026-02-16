package dto

import "CONVERDA/internal/apps/domain/model/entity"

type SystemEnvironmentResponse struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

func ToSystemEnvironmentResponse(e *entity.SystemEnvironment) *SystemEnvironmentResponse {
	if e == nil {
		return nil
	}
	return &SystemEnvironmentResponse{
		Code:        e.Code,
		Name:        e.Name,
		Description: e.Description,
	}
}

func ToSystemEnvironmentResponseList(envs []*entity.SystemEnvironment) []*SystemEnvironmentResponse {
	resp := make([]*SystemEnvironmentResponse, len(envs))
	for i, e := range envs {
		resp[i] = ToSystemEnvironmentResponse(e)
	}
	return resp
}
