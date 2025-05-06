package domain

type TestingTeam int

type Team struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	TeamRole int    `json:"team_role"`
}
