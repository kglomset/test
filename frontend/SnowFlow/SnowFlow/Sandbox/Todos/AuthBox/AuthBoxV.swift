import SwiftUI

struct AuthBoxV: View {
    @StateObject private var vm = AuthBoxVM()
    
    var body: some View {
        VStack(spacing: 0) {
            List {
                ForEach(Array(vm.todoItems.keys.sorted()), id: \.self) { id in
                    if let status = vm.todoItems[id] {
                        switch status {
                        case .loaded(let todo):
                            TodoRow(todo: todo)
                        case .loading:
                            LoadingRow(id: id)
                        case .error(let error):
                            ErrorRow(todoId: id, error: error, onRetry: { _ in vm.fetchRandomTodo() })
                        }
                    }
                }
            }
            Button(action: {
                vm.fetchRandomTodo()
            }) {
                Text("Fetch Random Todo")
                    .frame(maxWidth: .infinity)
                    .padding()
                    .foregroundColor(.white)
                    .background(Color.blue)
                    .cornerRadius(8)
            }
            .padding()
        }
    }
}
