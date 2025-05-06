import SwiftUI
import Combine

@MainActor
class TodoViewModel: ObservableObject {
    
    @Published var todoItems: [Int: TodoItemStatus] = [:]
    @Published var isInitialLoading = false
    @Published var manualFetchInProgress = false
    
    private var activeTokens: [Int: NetworkRequestToken] = [:]
    private var manualFetchToken: (id: Int, token: NetworkRequestToken)?
    
    private let todoService: TodoService
    private let testTodos = Array(1...20)
    
    init() {
        todoService = TodoService()
    }
    
    func loadTodos() {
        isInitialLoading = todoItems.isEmpty
        cancelAllRequests()
        
        // mark all as loading
        for id in testTodos {
            todoItems[id] = .loading
            loadSingleTodo(id: id)
        }
    }
    
    func refreshTodo(_ id: Int) {
        if let token = activeTokens[id] {
            Task { await todoService.cancelRequest(with: token) }
            activeTokens.removeValue(forKey: id)
        }
        todoItems[id] = .loading
        loadSingleTodo(id: id)
    }
    
    private func loadSingleTodo(id: Int) {
        let (token, task) = todoService.fetchTodo(id: id, delay: 1, errorRate: 20)
        activeTokens[id] = token
        
        Task {
            do {
                let todo = try await task.value
                todoItems[id] = .loaded(todo)
            } catch {
                todoItems[id] = .error(error)
            }
            
            activeTokens.removeValue(forKey: id)
            if !todoItems.values.contains(where: { $0.isLoading }) {
                isInitialLoading = false
            }
        }
    }
    
    private func cancelAllRequests() {
        for token in activeTokens.values {
            Task { await todoService.cancelRequest(with: token) }
        }
        activeTokens.removeAll()
    }
    
    func manualFetchTodo(id: Int, delay: Int, errorRate: Int) {
        guard !manualFetchInProgress else { return }
        manualFetchInProgress = true
        
        todoItems[id] = .loading
        let (token, task) = todoService.fetchTodo(id: id, delay: delay, errorRate: errorRate)
        manualFetchToken = (id, token)
        
        Task {
            do {
                let todo = try await task.value
                todoItems[id] = .loaded(todo)
            } catch {
                todoItems[id] = .error(error)
            }
            manualFetchToken = nil
            manualFetchInProgress = false
        }
    }
    
    func cancelManualFetch() {
        if let tokenData = manualFetchToken {
            Task { await todoService.cancelRequest(with: tokenData.token) }
            manualFetchToken = nil
            manualFetchInProgress = false
        }
    }
    
    var successfulTodos: [TodoM] {
        todoItems.compactMap { $0.value.todo }
    }
    
    func postTodo() {
        let (_, task) = todoService.postTodo()
        
        Task {
            do {
                let newTodo = try await task.value
                todoItems[newTodo.id] = .loaded(newTodo)
            } catch {
                print("Error posting todo: \(error)")
            }
        }
    }
}
