package config

import "Forum_BE/models"

type Permissions struct {
	Resource string
	Action   string
	Allowed  bool
}

var RolePermissions = map[models.Role][]Permissions{
	models.RoleRoot: {
		{Resource: "user", Action: "create", Allowed: true},
		{Resource: "user", Action: "edit", Allowed: true},
		{Resource: "user", Action: "delete", Allowed: true},
		{Resource: "question", Action: "create", Allowed: true},
		{Resource: "question", Action: "edit", Allowed: true},
		{Resource: "question", Action: "delete", Allowed: true},
		{Resource: "groups", Action: "create", Allowed: true},

		// Thêm các quyền khác cho root
	},
	models.RoleAdmin: {
		{Resource: "user", Action: "create", Allowed: true},
		{Resource: "user", Action: "edit", Allowed: true},
		{Resource: "user", Action: "delete", Allowed: true},
		{Resource: "question", Action: "create", Allowed: true},
		{Resource: "question", Action: "edit", Allowed: true},
		{Resource: "question", Action: "delete", Allowed: true},
		{Resource: "groups", Action: "create", Allowed: true},

		// Thêm các quyền khác cho admin
	},
	models.RoleEmployee: {
		{Resource: "question", Action: "create", Allowed: true},
		{Resource: "question", Action: "edit", Allowed: false},
		{Resource: "question", Action: "delete", Allowed: false},
		// Thêm các quyền khác cho employee
	},
	models.RoleUser: {
		{Resource: "question", Action: "create", Allowed: true},
		{Resource: "question", Action: "edit", Allowed: false},
		{Resource: "question", Action: "delete", Allowed: false},
		// Thêm các quyền khác cho user
	},
}
