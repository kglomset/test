import SwiftUI

@MainActor
class AuthBoxVM: ObservableObject {
    @Published var todoItems: [Int: TodoItemStatus] = [:]
    
    private let service = AuthTodoService()
    
    /// Fetches a random todo and updates the todoItems dictionary.
    func fetchRandomTodo() {
        Task {
            do {
                let todo = try await service.fetchRandomTodo()
                todoItems[todo.id] = .loaded(todo)
            } catch {
                // in case of an error, store the error under a random key
                
                let id = Int.random(in: 400...500)
                todoItems[id] = .error(error)
            }
        }
    }
}
