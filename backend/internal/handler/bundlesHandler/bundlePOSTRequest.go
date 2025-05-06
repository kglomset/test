package bundlesHandler

type BundlePOSTRequest struct {
	Products    []int  `json:"products"`
	ProductName string `json:"name"`
	Comment     string `json:"comment"`
	IsPublic    bool   `json:"is_public"`
	TestingTeam int    `json:"testing_team"`
	Status      string `json:"status"`
}
