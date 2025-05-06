import Foundation

class TestsViewModel: ObservableObject {
    @Published var sentence: String = "Do not change me!"
    @Published var inputSentence: String = ""
    
    /// Updates the sentence
    func updateText() {
        sentence = inputSentence
    }
}
