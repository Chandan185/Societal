package store

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PaginatedFeedQuery struct {
	Limit  int      `json:"limit" validate:"gte=1,lte=20"`
	Offset int      `json:"offset" validate:"gte=0"`
	Sort   string   `json:"sort" validate:"oneof=asc desc"`
	Tags   []string `json:"tags" validate:"max=5"`
	Search string   `json:"search" validate:"max=100"`
	Since  string   `json:"since"`
	Until  string   `json:"until"`
}

func (pq PaginatedFeedQuery) Parse(r *http.Request) (PaginatedFeedQuery, error) {
	qs := r.URL.Query()
	limit := qs.Get("limit")
	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return pq, err
		}
		pq.Limit = l
	}
	offset := qs.Get("offset")
	if offset != "" {
		o, err := strconv.Atoi(limit)
		if err != nil {
			return pq, err
		}
		pq.Offset = o
	}
	sort := qs.Get("sort")
	if sort != "" {
		pq.Sort = sort
	}
	tags := qs.Get("tags")
	if tags != "" {
		pq.Tags = strings.Split(tags, ",")
	}
	search := qs.Get("search")
	if search != "" {
		pq.Search = search
	}
	since := qs.Get("since")
	if since != "" {
		pq.Since = parseTime(since)
	}
	until := qs.Get("until")
	if until != "" {
		pq.Until = parseTime(until)
	}
	return pq, nil
}

func parseTime(s string) string {
	t, err := time.Parse(time.DateTime, s)
	if err != nil {
		return ""
	}
	return t.Format(time.DateTime)
}
