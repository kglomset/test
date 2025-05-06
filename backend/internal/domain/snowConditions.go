package domain

type SnowConditions struct {
	ID           int     `json:"id"`
	Temperature  float32 `json:"temperature"`
	SnowType     string  `json:"snow_type"`
	SnowHumidity string  `json:"snow_humidity"` // 'DS', 'W1', 'W2', 'W3', 'W4'
}
