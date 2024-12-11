package models

type Role string

const (
	RoleRoot     Role = "root"
	RoleAdmin    Role = "admin"
	RoleEmployee Role = "employee"
	RoleUser     Role = "user"
)
