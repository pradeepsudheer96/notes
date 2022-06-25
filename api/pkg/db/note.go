package db

import "context"

func (n *Note) Create(ctx context.Context, note *Note) (*Note, error) {
	stmt := `
		INSERT INTO notes(note)
		VALUES($1)
		RETURNING id, note;`
	var nn Note
	err := n.DB.QueryRowContext(ctx, stmt, note).Scan(&nn.ID, &nn.Title)
	if err != nil {
		return &Note{}, err
	}
	return &nn, nil
}
