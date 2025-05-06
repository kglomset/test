import Foundation

// MARK: - todo model
struct TodoM: Identifiable, Codable {
    let id: Int
    let title: String
    let completed: Bool
}

// MARK: - todo item status
enum TodoItemStatus {
    case loading
    case loaded(TodoM)
    case error(Error)
    
    var isLoading: Bool {
        if case .loading = self { return true }
        return false
    }
    
    var todo: TodoM? {
        if case .loaded(let todo) = self { return todo }
        return nil
    }
    
    var error: Error? {
        if case .error(let error) = self { return error }
        return nil
    }
}

