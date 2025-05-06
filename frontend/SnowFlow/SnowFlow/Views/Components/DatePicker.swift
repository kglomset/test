import SwiftUI

struct DatePickerView: View {
    @Binding var selectedDate: Date
    @State private var showDatePicker: Bool = false

    var body: some View {
        VStack(alignment: .leading, spacing: 2) {
            Text("Date")
                .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.bodySize))
                .fontWeight(.medium)
            Button {
                showDatePicker.toggle()
            } label: {
                HStack() {
                    Text(dateString)
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                // get from the core style instead
                .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.bodySize))
                .foregroundColor(Theme.Colors.primary)
                .padding(.horizontal, 16)
                .frame(height: 38)
                .background(Theme.Colors.backgroundLight)
                .cornerRadius(Theme.CornerRadius.medium)
                .overlay(
                    RoundedRectangle(cornerRadius: Theme.CornerRadius.medium)
                        .stroke(Theme.Colors.border, lineWidth: 1)
                )
            }
            .popover(isPresented: $showDatePicker, attachmentAnchor: .point(.bottom)) {
                popoverContent
            }
        }
    }

    private var popoverContent: some View {
        VStack {
            HStack {
                Spacer()
                Button("Close") {
                    showDatePicker = false
                }
                .padding(.horizontal)
                .foregroundColor(.blue)
            }
            DatePicker("Select date", selection: $selectedDate, displayedComponents: .date)
                .datePickerStyle(GraphicalDatePickerStyle())
                .padding()
        }
        
    }

    private var dateString: String {
        let formatter = DateFormatter()
        formatter.dateStyle = .long
        return formatter.string(from: selectedDate)
    }
}

#Preview {
    @Previewable @State var date = Date()
    @Previewable @State var text: String = ""
    HStack{
        TextInput(
            text: $text,
            label: "Normal field",
            placeholder: "Enter text",
            hasBorder: true,
            externalError: nil
        )
        DatePickerView(selectedDate: $date)
    }
}
