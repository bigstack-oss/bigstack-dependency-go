package influx

import (
	"fmt"
	"strings"
)

type Query struct {
	stages []string
}

func (q *Query) Bucket(params string) *Query {
	q.stages = append(q.stages, fmt.Sprintf(`from(bucket: "%s")`, params))
	return q
}

func (q *Query) Range(params string) *Query {
	q.stages = append(q.stages, fmt.Sprintf(`|> range(%s)`, params))
	return q
}

func (q *Query) Measurement(params string) *Query {
	q.stages = append(q.stages, fmt.Sprintf(`|> filter(fn: (r) => r._measurement == "%s")`, params))
	return q
}

func (q *Query) Filter(params string) *Query {
	q.stages = append(q.stages, fmt.Sprintf(`|> filter(fn: (r) => r.%s)`, params))
	return q
}

func (q *Query) Pivot(params string) *Query {
	q.stages = append(q.stages, fmt.Sprintf(`|> pivot(%s)`, params))
	return q
}

func (q *Query) Group(params string) *Query {
	if params == "" {
		q.stages = append(q.stages, "|> group()")
	} else {
		q.stages = append(q.stages, fmt.Sprintf(`|> group(%s)`, params))
	}

	return q
}

func (q *Query) Sort(params string) *Query {
	q.stages = append(q.stages, fmt.Sprintf(`|> sort(%s)`, params))
	return q
}

func (q *Query) Limit(params string) *Query {
	q.stages = append(q.stages, fmt.Sprintf(`|> limit(%s)`, params))
	return q
}

func (q *Query) Count(params string) *Query {
	q.stages = append(q.stages, fmt.Sprintf(`|> count(%s)`, params))
	return q
}

func (q *Query) Rename(params string) *Query {
	q.stages = append(q.stages, fmt.Sprintf(`|> rename(%s)`, params))
	return q
}

func (q *Query) Keep(params string) *Query {
	q.stages = append(q.stages, fmt.Sprintf(`|> keep(%s)`, params))
	return q
}

func (q *Query) Distinct(params string) *Query {
	q.stages = append(q.stages, fmt.Sprintf(`|> distinct(%s)`, params))
	return q
}

func (q *Query) String() string {
	return strings.Join(q.stages, " ")
}
