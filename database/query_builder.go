package database

import (
	"fmt"
	"strings"
)

type QueryBuilder struct {
	Links		[]string
	Keywords 	[]string
	DateFrom	string
	DateTo		string
	SortBy		string
	Limit		int	
}

/* filter for links */
func (qb *QueryBuilder) FilterLinks() string {
	var conditions []string
	for _, link := range qb.Links {
		conditions = append(conditions, fmt.Sprintf("html Like '%%%s%%'", link))
	}
	return strings.Join(conditions, " OR ")
}

/* filter for keywords */
func (qb *QueryBuilder) FilterKeywords() string {
    var conditions []string
    for _, kw := range qb.Keywords {
        conditions = append(conditions, fmt.Sprintf("html LIKE '%%%s%%'", kw))
    }
    return strings.Join(conditions, " OR ")
}

/* filter date */
func (qb *QueryBuilder) FilterDate() string {
    if qb.DateFrom != "" && qb.DateTo != "" {
        return fmt.Sprintf("scraped_at BETWEEN '%s' AND '%s'", qb.DateFrom, qb.DateTo)
    }
    return ""
}

/* sort by date & bytesize */
func (qb *QueryBuilder) Sort() string {
    switch qb.SortBy {
    case "date":
        return "ORDER BY scraped_at DESC"
    case "size":
        return "ORDER BY octet_length(html) DESC"
    default:
        return ""
    }
}
/* limit results */
func (qb *QueryBuilder) LimitClause() string {
    if qb.Limit > 0 {
        return fmt.Sprintf("LIMIT %d", qb.Limit)
    }
    return "LIMIT 100" // Default
}