import SwiftUI

struct NetworkCancellationTestView: View {
    @StateObject private var viewModel = NetworkCancellationViewModel()
    @State private var selectedTest: Int? = nil
    
    var body: some View {
        NavigationView {
            VStack {
                List {
                    Section(header: Text("Test scenarios")) {
                        Button("1. Cancel in-progress request") {
                            selectedTest = 1
                            viewModel.testCancelInProgressRequest()
                        }
                        .foregroundColor(.blue)
                        
                        Button("2. Cancel during retry cooldown") {
                            selectedTest = 2
                            viewModel.testCancelDuringRetryCooldown()
                        }
                        .foregroundColor(.blue)
                        
                        Button("3. Cancel queued request") {
                            selectedTest = 3
                            viewModel.testCancelQueuedRequest()
                        }
                        .foregroundColor(.blue)
                        
                        Button("4. Multiple requests with cancellation") {
                            selectedTest = 4
                            viewModel.testMultipleRequestsWithCancellation()
                        }
                        .foregroundColor(.blue)
                        
                        Button("5. Cancel already completed request") {
                            selectedTest = 5
                            viewModel.testCancelCompletedRequest()
                        }
                        .foregroundColor(.blue)
                        
                        Button("6. Cancel non-existent Request") {
                            selectedTest = 6
                            viewModel.testCancelNonExistentRequest()
                        }
                        .foregroundColor(.blue)
                    }
                    
                    Section(header: Text("Active requests")) {
                        if viewModel.activeRequests.isEmpty {
                            Text("No active requests")
                                .foregroundColor(.gray)
                                .italic()
                        } else {
                            ForEach(Array(viewModel.activeRequests), id: \.key) { id, url in
                                VStack(alignment: .leading) {
                                    Text(url)
                                        .font(.headline)
                                    Text(id.uuidString)
                                        .font(.caption)
                                        .foregroundColor(.gray)
                                }
                            }
                        }
                    }
                    
                    Section(header: HStack {
                        Text("Logs")
                        Spacer()
                        Button("Clear") {
                            viewModel.clearLogs()
                        }
                        .font(.caption)
                    }) {
                        if viewModel.logs.isEmpty {
                            Text("No logs yet")
                                .foregroundColor(.gray)
                                .italic()
                        } else {
                            ForEach(viewModel.logs.indices, id: \.self) { index in
                                Text(viewModel.logs[viewModel.logs.count - 1 - index])
                                    .font(.system(.caption, design: .monospaced))
                                    .padding(.vertical, 2)
                            }
                        }
                    }
                }
                .listStyle(InsetGroupedListStyle())
            }
            .navigationTitle("Network cancellation tests")
            .overlay(
                Group {
                    if viewModel.isLoading {
                        HStack {
                            ProgressView()
                                .progressViewStyle(CircularProgressViewStyle())
                            Text("Network activity")
                                .padding(.leading, 8)
                        }
                        .padding()
                        .background(Color(.systemBackground))
                        .cornerRadius(8)
                        .shadow(radius: 4)
                    }
                }
            )
        }
    }
}
