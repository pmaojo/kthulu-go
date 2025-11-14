package authz

# Basic example policy: allow if user has role in allowed_roles input
default allow = false

allow {
  some role
  role := input.user.role
  role == "admin"
}
