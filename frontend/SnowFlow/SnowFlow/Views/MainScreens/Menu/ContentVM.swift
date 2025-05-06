import SwiftUI

@MainActor
final class ContentViewModel: ObservableObject {
    @Published var showLogin = !SessionManagerProxy.shared.isAuthenticated

    init() {
        SessionManagerProxy.shared.$isAuthenticated
            .map { !$0 }
            .assign(to: &$showLogin)
    }
}

