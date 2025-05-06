import SwiftUI

struct ProfileView: View {
    @ObservedObject var viewModel: ProfileViewModel

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(alignment: .leading, spacing: 64) {
                    profileInfoSection
                    changePasswordSection
                    teamMembersSection
                    loggedInDevicesSection
                    appInfoNavigation
                }
                .padding(.horizontal, 16)
            }
        }
    }
}

// MARK: - Subviews
private extension ProfileView {
    var profileInfoSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(Strings.Profile.profileTitle)
                .font(.headline)
            
            labeledText(label: Strings.Profile.emailLabel, text: viewModel.userEmail)
            labeledText(label: "Team", text: viewModel.userRole)
            labeledText(label: Strings.Profile.roleLabel, text: viewModel.userRole)
        }
    }

    var changePasswordSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(Strings.Profile.changePassword)
                .font(.headline)
            
            PasswordInput(password: $viewModel.oldPassword, label: Strings.Profile.oldPassword, placeholder: Strings.Profile.oldPassword, hasBorder: false, externalError: nil)
            PasswordInput(password: $viewModel.newPassword, label: Strings.Profile.newPassword, placeholder: Strings.Profile.newPassword, hasBorder: false, externalError: nil)
            PasswordInput(password: $viewModel.confirmPassword, label: Strings.Profile.confirmPassword, placeholder: Strings.Profile.confirmPassword, hasBorder: false, externalError: viewModel.passwordError)
            
            PrimaryButton(title: Strings.Profile.updatePassword, expand: true) {
                viewModel.updatePassword()
            }
        }
    }

    var teamMembersSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(Strings.Profile.teamMembers)
                .font(.headline)
            
            ForEach(viewModel.teamMembers, id: \.id) { member in
                teamMemberRow(member: member)
            }
            
            inviteTeamMemberRow
        }
    }

    var loggedInDevicesSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(Strings.Profile.loggedInDevices)
                .font(.headline)
            
            ForEach(viewModel.devices, id: \.id) { device in
                loggedInDeviceRow(device: device)
            }
        }
    }

    var appInfoNavigation: some View {
        NavigationLink(destination: ContentView()) {
            PrimaryButton(title: Strings.Profile.appInfo, expand: true) {}
        }
        .padding(.top, 16)
    }

    // MARK: - Helper Views
    func labeledText(label: String, text: String) -> some View {
        VStack(alignment: .leading, spacing: 4) {
            Text(label)
                .font(.subheadline)
                .foregroundColor(.gray)
            Text(text)
                .font(.body)
        }
    }

    func teamMemberRow(member: TeamMember) -> some View {
        HStack {
            VStack(alignment: .leading) {
                Text(member.email)
                    .font(.body)
                Text(member.status)
                    .font(.caption)
                    .foregroundColor(.gray)
            }
            Spacer()
            Button(action: { viewModel.removeTeamMember(member) }) {
                Image(systemName: "trash")
                    .foregroundColor(.red)
            }
        }
    }

    var inviteTeamMemberRow: some View {
        HStack {
            TextField(Strings.Profile.inviteEmailPlaceholder, text: $viewModel.inviteEmail)
                .textFieldStyle(RoundedBorderTextFieldStyle())
            PrimaryButton(title: Strings.Profile.inviteButton, expand: false) {
                viewModel.inviteTeamMember()
            }
        }
    }

    func loggedInDeviceRow(device: Device) -> some View {
        HStack {
            Text("\(device.name)  \(device.ip)  Active: \(device.lastActive)")
                .font(.body)
            Spacer()
            Button(action: { viewModel.removeDevice(device) }) {
                Image(systemName: "trash")
                    .foregroundColor(.red)
            }
        }
    }
}

// MARK: - Preview
struct ProfileView_Previews: PreviewProvider {
    static var previews: some View {
        ProfileView(viewModel: ProfileViewModel())
    }
}
