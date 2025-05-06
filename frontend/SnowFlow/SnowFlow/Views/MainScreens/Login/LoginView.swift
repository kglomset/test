import SwiftUI
import Combine

struct LoginView: View {
    @StateObject private var vm = LoginViewModel()
    
    var onLoginSuccess: () -> Void
    
    var body: some View {
        ZStack{
            
            GeometryReader { geometry in
                ParallaxBackgroundView()
                    .frame(width: geometry.size.width)
            }
            .ignoresSafeArea()
            
            VStack(spacing: 12) {
                
                EmailInput(
                    email: $vm.email,
                    label: nil,
                    placeholder: "Epost",
                    hasBorder: false,
                    externalError: nil
                )
                
                PasswordInput(
                    password: $vm.password,
                    label: nil,
                    placeholder: "Passord",
                    hasBorder: false,
                    externalError: nil
                )
                
                HStack{
                    Text("Glemt passord?")
                        .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.bodySize))
                        .foregroundColor(Theme.Colors.secondary)
                }
                .frame(maxWidth: .infinity, alignment: .trailing)
                .padding(.top, -8)
                
                PrimaryButton(title: Strings.Login.loginButton, expand: true) {
                    vm.login {
                        onLoginSuccess()
                    }
                }
                .padding(.top, 4)
                
                if let errorMessage = vm.errorMessage {
                    Text(errorMessage)
                        .font(.custom(Theme.Fonts.bodyName, size: Theme.Fonts.bodySize))
                        .foregroundColor(Theme.Colors.error)
                }
            }
            .frame(maxWidth: 320, maxHeight: .infinity, alignment: .bottom)
            .padding(.bottom, 24)
        }
    }
}



#Preview {
    LoginView(onLoginSuccess: {})
}

struct ParallaxBackgroundView: View {
    var speed: Double = 420
    var maxOffset: CGFloat = 550
    @State private var logoOpacity: Double = 0
    
    var body: some View {
        TimelineView(.animation) { timeline in
            let time = timeline.date.timeIntervalSinceReferenceDate
            let progress = (time.truncatingRemainder(dividingBy: speed)) / speed
            let normalizedOscillation = (sin(progress * 2 * .pi) + 1) / 2
            let offsetX = normalizedOscillation * maxOffset
            
            ZStack {
                Image("snowy_mountain_full")
                    .resizable()
                    .aspectRatio(contentMode: .fill)
                    .scaleEffect(1.2)
                    .offset(y: 40)
                    .ignoresSafeArea()
                    .offset(x: offsetX)
                
                Image("login_title")
                    .foregroundColor(Theme.Colors.primary)
                    .offset(y: -192)
                    .opacity(logoOpacity)
                
                Image("snowy_mountain_mountain")
                    .resizable()
                    .aspectRatio(contentMode: .fill)
                    .ignoresSafeArea()
                
                LinearGradient(
                    gradient: Gradient(stops: [
                        .init(color: Color(hex: "000814").opacity(0), location: 0.0),
                        .init(color: Color(hex: "000814").opacity(1), location: 1.0)
                    ]),
                    startPoint: .top,
                    endPoint: .bottom
                )
                .frame(height: 520)
                .frame(maxHeight: .infinity, alignment: .bottom)
                .ignoresSafeArea(.all)
                .opacity(0.12)
            }
            .onAppear {
                withAnimation(.easeIn(duration: 0.36).delay(0.54)) {
                    logoOpacity = 1
                }
            }
        }
    }
}



#Preview {
    ParallaxBackgroundView()
}
