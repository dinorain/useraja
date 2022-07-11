package dto

import "github.com/dinorain/useraja/pkg/utils"

type FindUserResponseDto struct {
	Meta utils.PaginationMetaDto `json:"meta"`
	Data interface{}              `json:"data"`
}
