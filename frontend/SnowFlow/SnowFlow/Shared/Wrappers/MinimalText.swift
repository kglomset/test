// code from: https://medium.com/@brianmasse_94741/universaltext-perfecting-swiftui-text-9147228e9231
// modified to basicly only accomponate line spacing

// does not work with automatic line breaks...

import SwiftUI

enum ProvidedFont: String {
    case primary = "Bricolage Grotesque"
    case secondary = "Poppins"
}

struct MinimalText: View {
    let text: String
    let size: CGFloat
    let font: String
    let lineSpacing: CGFloat

    init(_ text: String, size: CGFloat, font: ProvidedFont = .primary, lineSpacing: CGFloat = 0.5) {
        self.text = text
        self.size = size
        self.font = font.rawValue
        self.lineSpacing = lineSpacing
    }

    @ViewBuilder
    private func renderText(_ text: String) -> some View {
        Text(text)
            .font(Font.custom(font, size: size))
            .lineSpacing(lineSpacing)
    }

    @ViewBuilder
    var body: some View {
        if lineSpacing < 0 {
            let texts = text.components(separatedBy: "\n")
            VStack(alignment: .leading, spacing: 0) {
                ForEach(0..<texts.count, id: \.self) { i in
                    renderText(texts[i])
                        .offset(y: CGFloat(i) * lineSpacing)
                }
            }
            .padding(.bottom, CGFloat(texts.count - 1) * lineSpacing)
        } else {
            renderText(text)
        }
    }
}

struct DisplayMinimalTextView: View {
    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            MinimalText("MinimalText Demo", size: 24, font: .primary)
                .bold()
                .foregroundColor(.blue)
            
            Divider()
            
            MinimalText("Hello World!\nNew line here", size: 18, font: .secondary)
                .foregroundColor(.purple)
            
            MinimalText("Increased spacing\nNew line here", size: 18, font: .secondary, lineSpacing: 10)
                .foregroundColor(.green)
            
            MinimalText("Reduced spacing\nNew line here", size: 18, font: .primary, lineSpacing: -5)
                .foregroundColor(.red)
            
            Divider()
        }
        .padding(16)
    }
}

#Preview {
    DisplayMinimalTextView()
}
