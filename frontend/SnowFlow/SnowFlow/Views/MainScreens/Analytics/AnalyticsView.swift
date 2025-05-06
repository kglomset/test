import SwiftUI

struct AnalyticsView: View {
    
    @ObservedObject var vm: AnalyticsViewModel
    
    var body: some View {
        NavigationView {
            ScrollView {
                VStack {
                    
                    PrimaryButton(title: "Do a network call", expand: false) {
                        vm.doNetworkCall()
                    }
                    
                }
            }
        }
    }
}
