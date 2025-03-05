package storage

import (
	"net/http"
	"strconv"
	"strings"
)

type PaginateQuery struct {
	Limit  int      `json:"limit" validate:"gte=1,lte=25"`
	Offset int      `json:"offset" validate:"gte=0"`
	Sort   string   `json:"sort" validate:"oneof=asc desc"`
	Search string   `json:"search" validate:"max=100"`
	Tags   []string `json:"tags" validate:"max=5"`
}


func (pq PaginateQuery) Parse(r *http.Request) (PaginateQuery, error) {
	queryS := r.URL.Query()
	limit := queryS.Get("limit")
	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return pq, err
		}
		pq.Limit = l
	}

	offset := queryS.Get("offset")
	if offset != "" {
		o, err := strconv.Atoi(offset)
		if err != nil {
			return pq, err
		}
		pq.Offset = o
	}

	sort := queryS.Get("sort")
	if sort != "" {

		pq.Sort = sort
	}

	search := queryS.Get("search")
	if search != "" {

		pq.Search = search
	}
	tags := queryS.Get("tags")
	if tags != "" {
		pq.Tags = strings.Split(tags, ",")
	}
	return pq, nil

}
