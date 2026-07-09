/**
 * @License OTMC License
 * @Copyright (c) 2026 OTMC Softwares. All rights reserved.
 * @Contributors Trung Ng, OTMC Authors.
**/

package db

import (
	"embed"
)

//go:embed schemas/*.sql
var SchemaFS embed.FS

//go:embed migration/pre/*.sql migration/post/*.sql migration/samples/*.sql
var MigrationFS embed.FS
