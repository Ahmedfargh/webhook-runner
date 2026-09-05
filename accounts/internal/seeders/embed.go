package seeders

import (
	_ "embed"
)

//go:embed countries.json
var CountriesJSON []byte

//go:embed permissions.json
var PermissionsJSON []byte

//go:embed roles.json
var RolesJSON []byte

//go:embed admins.json
var AdminsJSON []byte
