import SwiftUI

// MARK: - InputField (Core)
struct InputField: View {
    @Binding var text: String
    @FocusState.Binding var isFocused: Bool
    
    let label: String?
    let placeholder: String
    let isSecure: Bool
    let keyboardType: UIKeyboardType
    let textContentType: UITextContentType?
    let capitalization: TextInputAutocapitalization
    let hasBorder: Bool
    let errorMessage: String?
    
    private var borderColor: Color? {
        if let errorMessage = errorMessage, !errorMessage.isEmpty {
            return Theme.Colors.error // border error color when error exists
        } else if isFocused {
            return Theme.Colors.primary // highlight when focused
        } else {
            return hasBorder ? Theme.Colors.border : nil // default or no border
        }
    }
    
    var body: some View {
        VStack(alignment: .leading, spacing: 2) {
            if let label = label, !label.isEmpty {
                Text(label)
                    .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.bodySize))
                    .fontWeight(.medium)
            }
            
            Group {
                if isSecure {
                    SecureField(placeholder, text: $text)
                } else {
                    TextField(placeholder, text: $text)
                        .keyboardType(keyboardType)
                        .textContentType(textContentType)
                }
            }
            .modifier(InputFieldModifier(borderColor: borderColor))
            .autocorrectionDisabled(true)
            .textInputAutocapitalization(isSecure ? .never : capitalization)
            .focused($isFocused)
            
            if let errorMessage = errorMessage, !errorMessage.isEmpty {
                Text(errorMessage)
                    .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.captionSize))
                    .foregroundColor(Theme.Colors.error)
            }
        }
    }
}

// MARK: - InputField styling
struct InputFieldModifier: ViewModifier {
    let borderColor: Color?
    
    func body(content: Content) -> some View {
        let effectiveBorderColor = borderColor ?? Theme.Colors.backgroundLight
        
        return content
            .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.bodySize))
            .foregroundColor(Theme.Colors.primary)
            .padding(.horizontal, 16)
            .frame(height: 38)
            .background(Theme.Colors.backgroundLight)
            .cornerRadius(Theme.CornerRadius.medium)
            .overlay(
                RoundedRectangle(cornerRadius: Theme.CornerRadius.medium)
                    .stroke(effectiveBorderColor, lineWidth: 1)
            )
    }
}

// MARK: - TextInput
struct TextInput: View {
    @Binding var text: String
    @FocusState private var isFocused: Bool
    
    let label: String?
    let placeholder: String
    let hasBorder: Bool
    let externalError: String?
    
    var body: some View {
        InputField(
            text: $text,
            isFocused: $isFocused,
            label: label,
            placeholder: placeholder,
            isSecure: false,
            keyboardType: .default,
            textContentType: nil,
            capitalization: .sentences,
            hasBorder: hasBorder,
            errorMessage: externalError
        )
    }
}

// MARK: - EmailInput
struct EmailInput: View {
    @Binding var email: String
    @State private var validationError: String? = nil
    @FocusState private var isFocused: Bool
    
    let label: String?
    let placeholder: String
    let hasBorder: Bool
    let externalError: String?
    
    var body: some View {
        InputField(
            text: $email,
            isFocused: $isFocused,
            label: label,
            placeholder: placeholder,
            isSecure: false,
            keyboardType: .emailAddress,
            textContentType: .emailAddress,
            capitalization: .never,
            hasBorder: hasBorder,
            errorMessage: resolvedError()  // prioritize internal error
        )
        .onChange(of: isFocused) { _, isNowFocused in
            if !isNowFocused {
                validateEmail()  // validate when losing focus
            }
        }
        .onChange(of: email) { _, _ in
            if validationError != nil {  // only validate on typing if an error already exists
                validateEmail()
            }
        }
    }
    
    private func validateEmail() {
        let emailRegex = "^[A-Z0-9._%+-]+@[A-Z0-9.-]+\\.[A-Z]{2,}$"
        let predicate = NSPredicate(format: "SELF MATCHES[c] %@", emailRegex)
        validationError = predicate.evaluate(with: email) ? nil : "Invalid email format"
    }
    
    private func resolvedError() -> String? {
        return validationError ?? externalError  // show internal error first, then external
    }
}


// MARK: - PasswordInput
struct PasswordInput: View {
    @Binding var password: String
    @FocusState private var isFocused: Bool  // Added focus management
    
    let label: String?
    let placeholder: String
    let hasBorder: Bool
    let externalError: String?
    
    var body: some View {
        InputField(
            text: $password,
            isFocused: $isFocused,
            label: label,
            placeholder: placeholder,
            isSecure: true,
            keyboardType: .default,
            textContentType: .password,
            capitalization: .never,
            hasBorder: hasBorder,
            errorMessage: externalError
        )
    }
}

// MARK: - NumberInput
struct NumberInput: View {
    @Binding var numberText: String
    @State private var validationError: String? = nil
    @FocusState private var isFocused: Bool
    
    let label: String?
    let placeholder: String
    let hasBorder: Bool
    let externalError: String?
    let min: Double
    let max: Double
    
    var body: some View {
        InputField(
            text: $numberText,
            isFocused: $isFocused,
            label: label,
            placeholder: placeholder,
            isSecure: false,
            keyboardType: .decimalPad,
            textContentType: nil,
            capitalization: .never,
            hasBorder: hasBorder,
            errorMessage: resolvedError()
        )
        .onChange(of: isFocused) { _, isNowFocused in
            if !isNowFocused {
                validateNumber()  // validate on losing focus
            }
        }
        .onChange(of: numberText) { _, _ in
            if validationError != nil {
                validateNumber()  // re-validate while typing if an error exists
            }
        }
    }
    
    private func validateNumber() {
        guard !numberText.isEmpty else {
                validationError = nil  // no error if input is empty
                return
            }
        
        guard let number = Double(numberText) else {
            validationError = "Invalid number format"
            return
        }
        
        if number < min || number > max {
            validationError = "Number must be between \(min) and \(max)"
        } else {
            validationError = nil
        }
    }
    
    private func resolvedError() -> String? {
        return validationError ?? externalError  // show internal validation error first
    }
}

// MARK: - Preview
struct InputFieldsPreview: View {
    @State private var normalText1: String = ""
    @State private var normalText2: String = ""
    @State private var normalText3: String = ""
    @State private var numberText: String = ""
    @State private var emailText: String = ""
    @State private var secureText: String = ""
    
    var body: some View {
        VStack(spacing: Theme.Spacing.medium) {
            TextInput(
                text: $normalText1,
                label: "Normal Field 1",
                placeholder: "Enter text",
                hasBorder: false,
                externalError: nil
            )
            
            TextInput(
                text: $normalText2,
                label: nil,
                placeholder: "Enter text",
                hasBorder: true,
                externalError: nil
            )
            
            TextInput(
                text: $normalText3,
                label: "Normal field 3",
                placeholder: "Enter text",
                hasBorder: false,
                externalError: "Error: invalid input"
            )
            
            NumberInput(
                numberText: $numberText,
                label: "Number field",
                placeholder: "Enter number",
                hasBorder: true,
                externalError: nil,
                min: -50.0,
                max: 100.0
            )
            
            EmailInput(
                email: $emailText,
                label: "Email field",
                placeholder: "Enter email",
                hasBorder: false,
                externalError: nil
            )
            
            PasswordInput(
                password: $secureText,
                label: "Secure field",
                placeholder: "Enter password",
                hasBorder: false,
                externalError: nil
            )
        }
        .padding()
        .previewLayout(.sizeThatFits)
        .background(Theme.Colors.backgroundGray)
    }
}

struct InputFieldsPreview_Previews: PreviewProvider {
    static var previews: some View {
        InputFieldsPreview()
    }
}
