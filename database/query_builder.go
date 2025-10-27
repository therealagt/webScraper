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

/* all rawhtmlquery functions combined */
func (qb *QueryBuilder) BuildRawHTMLQuery() string {
    var whereClauses []string

    if links := qb.FilterLinks(); links != "" {
        whereClauses = append(whereClauses, links)
    }
    if keywords := qb.FilterKeywords(); keywords != "" {
        whereClauses = append(whereClauses, keywords)
    }
    if date := qb.FilterDate(); date != "" {
        whereClauses = append(whereClauses, date)
    }

    where := ""
    if len(whereClauses) > 0 {
        where = "WHERE " + strings.Join(whereClauses, " AND ")
    }

    query := fmt.Sprintf(
        "SELECT * FROM raw_html %s %s %s;",
        where,
        qb.Sort(),
        qb.LimitClause(),
    )
    return query
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
        return fmt.Sprintf("completed_at BETWEEN '%s' AND '%s'", qb.DateFrom, qb.DateTo)
    }
    return ""
}

/* sort by date & bytesize */
func (qb *QueryBuilder) Sort() string {
    switch qb.SortBy {
    case "date":
        return "ORDER BY completed_at DESC"
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
    return "LIMIT 100" 
}