package constants

import "github.com/MaroIshiku/dyniku/internal/models"

const (
	FAIL     models.Status = "failure"
	SUCCESS  models.Status = "success"
	UPTODATE models.Status = "up to date"
	UPDATING models.Status = "updating"
	UNSET    models.Status = "unset"
)
