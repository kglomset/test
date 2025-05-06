import Foundation

// TODO: - should share a prototype of Test ?

struct DraftM: Identifiable {
    let id: UUID
    let draftId: Int
    var title: String
    var date: Date?
    var testSamples: [SampleM] = []
    var track: String?
    var airTemp: Int?
    var airHumidity: Int?
    var snowTemp: Int?
    var snowType: String?
    var snowHardness: String?
    var snowMoisture: Int?
    var location: String?
    var weatherIcon: String?
    var isPrivate: Bool?
    var comment: String?

    init(
        id: UUID = UUID(),
        draftId: Int,
        title: String = "Untitled draft",
        date: Date? = nil,
        testSamples: [SampleM] = [],
        track: String? = nil,
        airTemp: Int? = nil,
        airHumidity: Int? = nil,
        snowTemp: Int? = nil,
        snowType: String? = nil,
        snowHardness: String? = nil,
        snowMoisture: Int? = nil,
        location: String? = nil,
        weatherIcon: String? = nil,
        isPrivate: Bool? = nil,
        comment: String? = nil
    ) {
        self.id = id
        self.draftId = draftId
        self.title = title
        self.date = date
        self.testSamples = testSamples
        self.track = track
        self.airTemp = airTemp
        self.airHumidity = airHumidity
        self.snowTemp = snowTemp
        self.snowType = snowType
        self.snowHardness = snowHardness
        self.snowMoisture = snowMoisture
        self.location = location
        self.weatherIcon = weatherIcon
        self.isPrivate = isPrivate
        self.comment = comment
    }
}
