package mapper

import (
	"auth-service/domain"
	"auth-service/infrastructure/dto"
)

func MapUserToAgentInformationDto(users []domain.User) []dto.AgentInformationDto {
	var agents []dto.AgentInformationDto

	for _, it := range users {
		agent := dto.AgentInformationDto{
			Username: it.Username,
			Email: it.Email,
			Name: it.Name,
			Surname: it.Surname,
			Web: it.Web,
		}

		agents = append(agents, agent)

	}
	return agents

}
