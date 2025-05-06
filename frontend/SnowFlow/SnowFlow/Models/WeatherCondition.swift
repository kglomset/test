import Foundation

struct WeatherCondition: Identifiable, SelectableItem {
    let id = UUID()
    let name: String
    let icon: String?
}

let weatherConditions: [WeatherCondition] = [
    WeatherCondition(name: "Clear sky", icon: "clearsky_day"),
    WeatherCondition(name: "Fair", icon: "fair_day"),
    WeatherCondition(name: "Partly cloudy", icon: "partlycloudy_day"),
    WeatherCondition(name: "Cloudy", icon: "cloudy"),
    WeatherCondition(name: "Fog", icon: "fog"),
    WeatherCondition(name: "Rain", icon: "rain"),
    WeatherCondition(name: "Sleet", icon: "sleet"),
    WeatherCondition(name: "Light snow", icon: "lightsnow"),
    WeatherCondition(name: "Snow", icon: "snow"),
    WeatherCondition(name: "Heavy snow", icon: "heavysnow")
]
