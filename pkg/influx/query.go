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
	q.stages = append(q.stages, fmt.Sprintf(`|> filter(%s)`, params))
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
	if params == "" {
		q.stages = append(q.stages, "|> count()")
	} else {
		q.stages = append(q.stages, fmt.Sprintf(`|> count(%s)`, params))
	}

	return q
}

func (q *Query) Different() *Query {
	q.stages = append(q.stages, "|> difference()")
	return q
}

func (q *Query) AggregateWindow(params string) *Query {
	q.stages = append(q.stages, fmt.Sprintf(`|> aggregateWindow(%s)`, params))
	return q
}

func (q *Query) Map(params string) *Query {
	q.stages = append(q.stages, fmt.Sprintf(`|> map(%s)`, params))
	return q
}

func (q *Query) Last() *Query {
	q.stages = append(q.stages, "|> last()")
	return q
}

func (q *Query) Derivative(params string) *Query {
	q.stages = append(q.stages, fmt.Sprintf(`|> derivative(%s)`, params))
	return q
}

func (q *Query) Max(params string) *Query {
	q.stages = append(q.stages, fmt.Sprintf(`|> max(%s)`, params))
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

func (q *Query) Top(params string) *Query {
	q.stages = append(q.stages, fmt.Sprintf(`|> top(%s)`, params))
	return q
}

func (q *Query) String() string {
	return strings.Join(q.stages, " ")
}
