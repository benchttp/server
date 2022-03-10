package postgresql

import (
	"github.com/lib/pq"

	"github.com/benchttp/server/benchttp"
)

func (s StatsService) ListAvailable(userID string) ([]benchttp.StatsDescriptor, error) {
	statsDescriptorList := []benchttp.StatsDescriptor{}

	stmt, err := s.db.Prepare(`SELECT id, tag, finished_at FROM stats_descriptor WHERE user_id = $1 ORDER BY finished_at DESC`)
	if err != nil {
		return []benchttp.StatsDescriptor{}, ErrPreparingStmt
	}
	defer stmt.Close()

	rows, err := stmt.Query(userID)
	if err != nil {
		return []benchttp.StatsDescriptor{}, ErrExecutingPreparedStmt
	}
	defer rows.Close()

	for rows.Next() {
		statsDescriptor := benchttp.StatsDescriptor{}
		err = rows.Scan(
			&statsDescriptor.ID,
			&statsDescriptor.Tag,
			&statsDescriptor.FinishedAt,
		)
		if err != nil {
			return nil, err
		}
		statsDescriptorList = append(statsDescriptorList, statsDescriptor)
	}

	return statsDescriptorList, nil
}

func (s StatsService) GetByID(statsDescriptorID string) (benchttp.Stats, error) {
	stats := benchttp.Stats{}

	stmt := `
SELECT
	s.id,
	s.tag,
	s.finished_at,
	c.code_1xx,
	c.code_2xx,
	c.code_3xx,
	c.code_4xx,
	c.code_5xx,
	t.min,
	t.max,
	t.mean,
	t.median,
	t.variance,
	t.deciles
FROM public.stats_descriptor AS s
INNER JOIN public.codestats AS c ON c.stats_descriptor_id = s.id
INNER JOIN public.timestats AS t ON t.stats_descriptor_id = s.id
WHERE s.id = $1
ORDER BY s.finished_at DESC`[1:]

	row := s.db.QueryRow(stmt, statsDescriptorID)
	err := row.Scan(
		&stats.StatsDescriptor.ID,
		&stats.StatsDescriptor.Tag,
		&stats.StatsDescriptor.FinishedAt,
		&stats.Codestats.Code1xx,
		&stats.Codestats.Code2xx,
		&stats.Codestats.Code3xx,
		&stats.Codestats.Code4xx,
		&stats.Codestats.Code5xx,
		&stats.Timestats.Min,
		&stats.Timestats.Max,
		&stats.Timestats.Mean,
		&stats.Timestats.Median,
		&stats.Timestats.Variance,
		(*pq.Float64Array)(&stats.Timestats.Deciles),
	)
	if err != nil {
		return stats, ErrScanningRows
	}

	return stats, nil
}