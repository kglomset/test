import SwiftUI

// MARK: - View model
class DetailedDraftVM: ObservableObject {
    
    @Published var emailText: String = ""
    
    private let initialDraftState: DraftM
    
    var title: String
    var titleError: String? = nil
    
    var date: Date
    var dateError: String? = nil
    
    var location: String
    var locationError: String? = nil
    
    @Published var airTemperature: String
    var airTemperatureError: String? = nil
    
    @Published var airHumidity: String
    var airHumidityError: String? = nil
    
    @Published var snowType: String
    
    @Published var snowHardness: String
    
    @Published var snowTemp: String
    var snowTempError: String? = nil
    
    @Published var snowHumidity: String
    var snowHumidityError: String? = nil
    
    @Published var weatherIcon: String
    
    var isPrivate: Bool
    
    var comment: String
    var commentError: String? = nil
    
    @Published var samples: [SampleM] = [] {
        didSet {
            products = samples.compactMap { $0.productPreview }
        }
    }
    @Published var products: [ProductPreviewM] = []
    
    var productsCount: Int {
        samples.count
    }
    
    var testCount: Int {
        samples.count - 1
    }
    
    // MARK: - Init
    init(id: Int) {
        let fetchedDraft = PredefinedDrafts.getDraft(by: id) // fetch from backend/storage using id
        
        self.initialDraftState = fetchedDraft
        self.title = fetchedDraft.title
        self.date = fetchedDraft.date ?? Date()
        self.location = fetchedDraft.location ?? ""
        self.airTemperature = fetchedDraft.airTemp.map { "\($0)" } ?? ""
        self.airHumidity = fetchedDraft.airHumidity.map { "\($0)" } ?? ""
        self.snowType = fetchedDraft.snowType ?? ""
        self.snowHardness = fetchedDraft.snowHardness ?? ""
        self.snowTemp = fetchedDraft.snowTemp.map { "\($0)" } ?? ""
        self.snowHumidity = fetchedDraft.snowMoisture.map { "\($0)" } ?? ""
        self.weatherIcon = fetchedDraft.weatherIcon ?? ""
        self.isPrivate = fetchedDraft.isPrivate ?? false
        self.comment = fetchedDraft.comment ?? ""
        self.samples = fetchedDraft.testSamples
    }
    
    // MARK: - Actions
    func saveDraft() {
        guard isDraftValid() else { return }
        // call save draft logic
    }
    
    func hasUnsavedChanges() -> Bool {
        return title != initialDraftState.title ||
        date != initialDraftState.date ||
        location != initialDraftState.location ||
        airTemperature != (initialDraftState.airTemp.map { "\($0)" } ?? "") ||
        airHumidity != (initialDraftState.airHumidity.map { "\($0)" } ?? "") ||
        snowType != (initialDraftState.snowType ?? "") ||
        snowHardness != (initialDraftState.snowHardness ?? "") ||
        snowTemp != (initialDraftState.snowTemp.map { "\($0)" } ?? "") ||
        snowHumidity != (initialDraftState.snowMoisture.map { "\($0)" } ?? "") ||
        weatherIcon != initialDraftState.weatherIcon ||
        isPrivate != initialDraftState.isPrivate ||
        comment != initialDraftState.comment //||
        //productsChanged()
    }
    
    private func isDraftValid() -> Bool {
        var isValid = true
        
        if title.trimmingCharacters(in: .whitespaces).isEmpty {
            titleError = "Title is required."
            isValid = false
        }
        
        if Int(airTemperature.trimmingCharacters(in: .whitespaces)) == nil {
            airTemperatureError = "Air temperature must be a number."
            isValid = false
        }
        
        return isValid
    }
    
    private func isValidSample(_ sample: SampleM) -> Bool {
        // check for nil values
        guard sample.isValid,
              let skiName = sample.skiName,
              let productPreview = sample.productPreview else {
            return false
        }
        
        // check for duplicates
        if samples.contains(where: { $0.skiName == skiName }) ||
            samples.contains(where: { $0.productPreview?.productId == productPreview.productId }) {
            return false
        }
        
        return true
    }
    
    func addSample() {
        samples.append(SampleM())
    }
    
    func removeSample(_ sample: SampleM) {
        samples.removeAll { $0.id == sample.id }
    }
}
