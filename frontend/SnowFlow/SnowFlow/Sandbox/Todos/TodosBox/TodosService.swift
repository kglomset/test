import Foundation
import Combine

class TodoService {
    private let networkManager: UltimateNetworkManager
    private let baseUrl = "http://localhost:9988"
    
    init() {
        let config = UltimateNetworkConfig(
            maxConcurrentLoads: 2,
            maxRetries: 2,
            retryDelay: 1.8,
            defaultHeaders: ["Service-Class": "network/box"],
            defaultTimeout: 12
        )
        networkManager = NetworkManagerFactory.create(config: config)
    }
    
    // helper method to construct the URL
    private func constructURL(id: Int, delay: Int, errorRate: Int) -> URL {
        var urlComponents = "\(baseUrl)"
        
        if delay > 0 {
            urlComponents += "/delay/\(delay)"
        }
        
        if errorRate > 0 {
            urlComponents += "/error/\(errorRate)"
        }
        
        urlComponents += "/todos/\(id)"
        
        guard let url = URL(string: urlComponents) else {
            fatalError("Invalid URL: \(urlComponents)")
        }
        return url
    }
    
    func fetchTodo(
        id: Int,
        delay: Int = 0,
        errorRate: Int = 0
    ) -> (NetworkRequestToken, Task<TodoM, Error>) {
        let url = constructURL(id: id, delay: delay, errorRate: errorRate)
        
        // initiate the network task, returns Data
        let (token, networkTask) = networkManager.fetch(
            url: url,
            method: "GET",
            timeout: 10.0,
            headers: ["Content-Type": "application/json"]
        )
        
        // create a new Task that decodes the network response data
        let decodingTask = Task<TodoM, Error> {
            // wait for the network task to finish
            let data = try await networkTask.value
            
            // decode the data into a TodoM directly (throws on failure)
            let todo: TodoM = try CodecManager.decode(data)
            return todo
        }
        
        return (token, decodingTask)
    }
    
    func cancelRequest(with token: NetworkRequestToken) async {
        await networkManager.cancelRequest(with: token)
    }
    
    private static var nextTodoId: Int = 200
    
    func postTodo() -> (NetworkRequestToken, Task<TodoM, Error>) {
        let id = TodoService.nextTodoId
        TodoService.nextTodoId += 1
        
        let randomNumber = Int.random(in: 20...80)
        let title = "Solve \(randomNumber) problems"
        let completed = Bool.random()
        
        let todo = TodoM(id: id, title: title, completed: completed)
        
        guard let url = URL(string: "\(baseUrl)/post") else {
            fatalError("Invalid URL for post endpoint")
        }
        
        // encode the todo to json
        let jsonData: Data
        do {
            jsonData = try JSONEncoder().encode(todo)
        } catch {
            fatalError("Failed to encode todo: \(error)")
        }
        
        // send the request
        let (token, networkTask) = networkManager.fetch(
            url: url,
            method: "POST",
            timeout: 8.0,
            headers: ["Content-Type": "application/json"],
            body: jsonData
        )
        
        // create a decoding task to extract the TodoM
        let decodingTask = Task<TodoM, Error> {
            let data = try await networkTask.value
            return try CodecManager.decode(data)
        }
        
        return (token, decodingTask)
    }
}
