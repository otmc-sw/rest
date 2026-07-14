/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/

package db

import (
	"embed"
)

//go:embed schemas/*.sql
var SchemaFS embed.FS

//go:embed migration/pre/*.sql migration/post/*.sql migration/samples/*.sql
var MigrationFS embed.FS
