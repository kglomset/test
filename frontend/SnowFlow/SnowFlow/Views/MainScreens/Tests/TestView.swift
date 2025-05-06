import SwiftUI

struct TestsView: View {
    
    @ObservedObject var viewModel: TestsViewModel
    
    var body: some View {
        NavigationView {
            ScrollView {
                VStack {
                    
                    Text("Tests")
                        .font(.largeTitle)
                        .bold()
                    
                    Spacer()
                    
                    Text(viewModel.sentence)
                    
                    Spacer()
                    
                    TextInput(
                        text: $viewModel.inputSentence,
                        label: "This is a input field",
                        placeholder: "Naturally you must enter text here",
                        hasBorder: false,
                        externalError: nil
                    )
                    
                    PrimaryButton(title: "Update text", expand: false) {
                        viewModel.updateText()
                    }
                    .padding(.top, 12)
                    
                    
                    Spacer()
                    
                }
            }
        }
    }
}
