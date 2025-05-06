import SwiftUI

// MARK: - Selection input component
struct SelectionInputView: View {
    let title: String
    let options: [(String, String?)] // tuple of (option, ssset name)
    let placeholder: String?
    @Binding var selectedOption: String
    
    var selectedIcon: String? {
        options.first { $0.0 == selectedOption }?.1
    }
    
    var displayText: String {
        selectedOption.isEmpty ? (placeholder ?? "") : selectedOption
    }
    
    var body: some View {
        VStack(alignment: .leading, spacing: 2) {
            Text(title)
                .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.bodySize))
                .fontWeight(.medium)
            
            Menu {
                ForEach(options, id: \.0) { option, icon in
                    Button(action: {
                        selectedOption = option
                    }) {
                        HStack {
                            if let icon = icon {
                                Image(icon)
                                    .resizable()
                                    .scaledToFit()
                                    .frame(width: 20, height: 20)
                            }
                            Text(option)
                        }
                    }
                }
            } label: {
                HStack(spacing: 2) {
                    if let icon = selectedIcon, !selectedOption.isEmpty {
                        Image(icon)
                            .resizable()
                            .scaledToFit()
                            .frame(width: 28, height: 28)
                    }
                    
                    Text(displayText)
                        .foregroundColor(selectedOption.isEmpty ? Color(UIColor.placeholderText) : Theme.Colors.primary)
                    
                    Spacer()
                    Image(systemName: "chevron.down")
                        .foregroundColor(Theme.Colors.border)
                }
                // get from the core style instead
                .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.bodySize))
                //.foregroundColor(Theme.Colors.primary)
                .padding(.horizontal, 16)
                .frame(height: 38)
                .background(Theme.Colors.backgroundLight)
                .cornerRadius(Theme.CornerRadius.medium)
                .overlay(
                    RoundedRectangle(cornerRadius: Theme.CornerRadius.medium)
                        .stroke(Theme.Colors.border, lineWidth: 1)
                )
            }
        }
    }
}

// MARK: - Wrapper for SelectText (text only)
struct SelectText: View {
    let title: String
    let options: [String]
    let placeholder: String?
    @Binding var selectedOption: String
    
    init(title: String, options: [String], placeholder: String? = nil, selectedOption: Binding<String>) {
        self.title = title
        self.options = options
        self.placeholder = placeholder
        self._selectedOption = selectedOption
    }
    
    var body: some View {
        SelectionInputView(
            title: title,
            options: options.map { ($0, nil) },
            placeholder: placeholder,
            selectedOption: $selectedOption
        )
    }
}

// MARK: - Wrapper for SelectTextIcon (with icons)
struct SelectTextIcon<Item: SelectableItem>: View {
    let title: String
    let items: [Item]
    let placeholder: String?
    @Binding var selectedOption: String
    
    init(title: String, items: [Item], placeholder: String? = nil, selectedOption: Binding<String>) {
        self.title = title
        self.items = items
        self.placeholder = placeholder
        self._selectedOption = selectedOption
    }
    
    var body: some View {
        SelectionInputView(
            title: title,
            options: items.map { ($0.name, $0.icon) },
            placeholder: placeholder,
            selectedOption: $selectedOption
        )
    }
}

// MARK: - Protocol for selectable items
protocol SelectableItem {
    var name: String { get }
    var icon: String? { get }
}

// MARK: - Preview
#Preview {
    @Previewable @State var selectedCity = ""
    @Previewable @State var selectedWeather = ""

    return VStack {
        SelectText(title: "Select City", options: ["Oslo", "New York", "Tokyo", "Berlin"], placeholder: "Choose a city", selectedOption: $selectedCity)
        
        SelectTextIcon(title: "Weather", items: weatherConditions, placeholder: "Choose weather", selectedOption: $selectedWeather)
    }
    .padding()
}
