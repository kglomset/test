import Foundation

struct Strings {
    
    struct Login {
        static let emailLabel = NSLocalizedString("email_label", comment: "Email field label")
        static let emailPlaceholder = NSLocalizedString("email_placeholder", comment: "Enter email placeholder")
        static let passwordLabel = NSLocalizedString("password_label", comment: "Password field label")
        static let passwordPlaceholder = NSLocalizedString("password_placeholder", comment: "Enter password placeholder")
        static let loginButton = NSLocalizedString("login_button", comment: "Login button title")
    }
    
    struct Profile {
        static let profileTitle = NSLocalizedString("profile_title", comment: "Profile Information title")
        static let emailLabel = NSLocalizedString("email_label", comment: "Email field label")
        static let roleLabel = NSLocalizedString("role_label", comment: "User role label")
        static let changePassword = NSLocalizedString("change_password", comment: "Change Password title")
        static let oldPassword = NSLocalizedString("old_password", comment: "Old Password field label")
        static let newPassword = NSLocalizedString("new_password", comment: "New Password field label")
        static let confirmPassword = NSLocalizedString("confirm_password", comment: "Confirm New Password field label")
        static let updatePassword = NSLocalizedString("update_password", comment: "Update Password button title")
        static let loggedInDevices = NSLocalizedString("logged_in_devices", comment: "Logged in devices title")
        static let appInfo = NSLocalizedString("app_info", comment: "App Information button title")
        static let teamMembers = NSLocalizedString("team_members", comment: "Team Members title")
        static let inviteEmailPlaceholder = NSLocalizedString("invite_email_placeholder", comment: "Enter team member email placeholder")
        static let inviteButton = NSLocalizedString("invite_button", comment: "Invite Team Member button title")
    }
    
    struct Common {
        static let error = NSLocalizedString("error", comment: "Generic error message")
    }
}
