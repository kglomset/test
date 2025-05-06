import SwiftUI

// MARK: - todo list view
struct TodoListView: View {
    @StateObject private var viewModel = TodoViewModel()
    
    var body: some View {
        VStack(spacing: 0) {
            TodoList(viewModel: viewModel)
                .task { viewModel.loadTodos() }
            
            ManualFetchView(viewModel: viewModel)
        }
        .frame(maxHeight: .infinity, alignment: .bottom)
    }
}

// MARK: - todo list
struct TodoList: View {
    @ObservedObject var viewModel: TodoViewModel
    
    var body: some View {
        List {
            ForEach(Array(viewModel.todoItems.keys.sorted()), id: \.self) { id in
                if let status = viewModel.todoItems[id] {
                    switch status {
                    case .loaded(let todo):
                        TodoRow(todo: todo)
                    case .loading:
                        LoadingRow(id: id)
                    case .error(let error):
                        ErrorRow(todoId: id, error: error, onRetry: viewModel.refreshTodo)
                    }
                }
            }
        }
        .refreshable { viewModel.loadTodos() }
    }
}

// MARK: - todo row
struct TodoRow: View {
    let todo: TodoM
    
    var body: some View {
        HStack {
            Text(todo.title)
                .font(.body)
            Spacer()
            Image(systemName: todo.completed ? "checkmark.circle.fill" : "circle")
                .foregroundColor(todo.completed ? .green : .gray)
        }
    }
}

// MARK: - loading row
struct LoadingRow: View {
    let id: Int
    
    var body: some View {
        HStack {
            Text("Todo \(id)")
                .font(.body)
                .foregroundColor(.gray)
            Spacer()
            ProgressView()
                .controlSize(.small)
        }
    }
}

// MARK: - error row
struct ErrorRow: View {
    let todoId: Int
    let error: Error
    let onRetry: (Int) -> Void
    
    var body: some View {
        HStack {
            Image(systemName: "exclamationmark.triangle")
                .foregroundColor(.orange)
            VStack(alignment: .leading) {
                Text("Todo \(todoId)")
                    .font(.subheadline)
                    .foregroundColor(.secondary)
                Text(error.localizedDescription)
                    .font(.caption)
                    .foregroundColor(.secondary)
            }
            Spacer()
            Button(action: { onRetry(todoId) }) {
                Image(systemName: "arrow.clockwise")
                    .foregroundColor(.blue)
            }
        }
    }
}

// MARK: - manual fetch view
struct ManualFetchView: View {
    @ObservedObject var viewModel: TodoViewModel
    
    @State private var idText = ""
    @State private var delayText = ""
    @State private var errorRateText = ""
    
    // validate that all inputs are non-empty and within their ranges
    // should be in vm, but well well this is only test code
    var isInputValid: Bool {
        if let id = Int(idText), id > 0,
           let delay = Int(delayText), delay >= 0, delay <= 5,
           let errorRate = Int(errorRateText), errorRate >= 0, errorRate <= 100 {
            return true
        }
        return false
    }
    
    var body: some View {
        VStack(spacing: 12) {
            HStack(spacing: 8) {
                TextField("ID", text: $idText)
                    .keyboardType(.numberPad)
                    .textFieldStyle(RoundedBorderTextFieldStyle())
                    .disabled(viewModel.manualFetchInProgress)
                TextField("Delay (0-5)", text: $delayText)
                    .keyboardType(.numberPad)
                    .textFieldStyle(RoundedBorderTextFieldStyle())
                    .disabled(viewModel.manualFetchInProgress)
                TextField("Error Rate (0-100)", text: $errorRateText)
                    .keyboardType(.numberPad)
                    .textFieldStyle(RoundedBorderTextFieldStyle())
                    .disabled(viewModel.manualFetchInProgress)
            }
            
            // fetch a specific todo
            Button(action: {
                if viewModel.manualFetchInProgress {
                    viewModel.cancelManualFetch()
                } else if isInputValid,
                          let id = Int(idText),
                          let delay = Int(delayText),
                          let errorRate = Int(errorRateText) {
                    viewModel.manualFetchTodo(id: id, delay: delay, errorRate: errorRate)
                    idText = ""; delayText = ""; errorRateText = ""
                }
            }) {
                Text(viewModel.manualFetchInProgress ? "Cancel" : "Fetch")
                    .frame(maxWidth: .infinity)
                    .padding()
                    .foregroundColor(.white)
                    .background(viewModel.manualFetchInProgress ? Color.red : Color.blue)
                    .cornerRadius(8)
            }
            .background(Color.yellow)
            .disabled(!isInputValid && !viewModel.manualFetchInProgress)
            
            // post a random todo
            Button(action: {
                viewModel.postTodo()
            }) {
                Text("Post a random todo")
                    .frame(maxWidth: .infinity)
                    .padding()
                    .foregroundColor(.white)
                    .background(Color.green)
                    .cornerRadius(8)
            }
        }
        .padding()
    }
}
