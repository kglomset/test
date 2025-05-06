package domain

type AirConditions struct {
	ID          int     `json:"id"`
	Temperature float32 `json:"temperature"`
	Humidity    int     `json:"humidity"`
	Wind        string  `json:"wind"`  //'S', 'L', 'M', 'ST'
	Cloud       string  `json:"cloud"` //'1', '2', '3', '4'
}
