import Foundation

class AnalyticsViewModel: ObservableObject {
    @Published var sentence: String = "Do not change me!"
    @Published var inputSentence: String = ""
    
    private var snowflowService = SnowFlowService()
    
    /// Updates the sentence
    func updateText() {
        sentence = inputSentence
    }
    
    func doNetworkCall() {
        Task {
            try await snowflowService.getTempTodo()
        }
    }
}
