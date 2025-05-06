import Foundation

struct Device: Identifiable {
    let id: UUID
    let name: String
    let ip: String
    let lastActive: String
}

struct TeamMember: Identifiable {
    let id: UUID
    let email: String
    let status: String
}

class ProfileViewModel: ObservableObject {
    @Published var userEmail: String = "post@waxmafia.biathlon"
    @Published var userRole: String = "Researcher"
    
    @Published var oldPassword: String = ""
    @Published var newPassword: String = ""
    @Published var confirmPassword: String = ""
    @Published var passwordError: String? = nil
    
    @Published var devices: [Device] = [
        Device(id: UUID(), name: "iPhone 12", ip: "172.217.22.14", lastActive: "12.01.25"),
        Device(id: UUID(), name: "iPhone 12", ip: "172.217.22.14", lastActive: "12.01.25"),
        Device(id: UUID(), name: "iPhone 12", ip: "172.217.22.14", lastActive: "12.01.25")
    ]
    
    @Published var teamMembers: [TeamMember] = [
        TeamMember(id: UUID(), email: "teammember1@example.com", status: "Active"),
        TeamMember(id: UUID(), email: "teammember2@example.com", status: "Pending")
    ]
    
    @Published var inviteEmail: String = ""
    
    @MainActor
    func updatePassword() {
        guard !oldPassword.isEmpty, !newPassword.isEmpty, !confirmPassword.isEmpty else {
            passwordError = "All fields are required"
            return
        }
        
        guard newPassword == confirmPassword else {
            passwordError = "Passwords do not match"
            return
        }
        
        // Simulate API call
        DispatchQueue.main.asyncAfter(deadline: .now() + 1) {
            self.oldPassword = ""
            self.newPassword = ""
            self.confirmPassword = ""
            self.passwordError = nil
        }
    }
    
    func removeDevice(_ device: Device) {
        devices.removeAll { $0.id == device.id }
    }
    
    func removeTeamMember(_ member: TeamMember) {
        teamMembers.removeAll { $0.id == member.id }
    }
    
    func inviteTeamMember() {
        guard !inviteEmail.isEmpty else { return }
        
        // Simulate inviting a new team member
        let newMember = TeamMember(id: UUID(), email: inviteEmail, status: "Pending")
        teamMembers.append(newMember)
        inviteEmail = ""
    }
}
