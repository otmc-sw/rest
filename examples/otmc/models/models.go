/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package models

// Role is automatically migrated to a "roles" table and exposed at /api/roles.
type Role struct {
	ID          uint64 `json:"id" otmc:"primary;auto"`
	Name        string `json:"name" otmc:"required;unique"`
	Description string `json:"description"`
}

// User is automatically migrated to a "users" table and exposed at /api/users.
// RoleID + Role give GORM a belongs-to relation to Role.
type User struct {
	ID     uint64 `json:"id" otmc:"primary;auto"`
	Name   string `json:"name" otmc:"required"`
	Email  string `json:"email" otmc:"required;unique"`
	RoleID uint64 `json:"role_id"`
	Role   *Role  `json:"role,omitempty"`
}
