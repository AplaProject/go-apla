package model

type Install struct {
	Progress string `gorm:not null;size:10`
}
