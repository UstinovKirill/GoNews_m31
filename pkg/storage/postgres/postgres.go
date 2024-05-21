package postgres

import (
	"context"
	"module_31/pkg/storage"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// Хранилище данных
type Store struct {
	db *sqlx.DB
}

// Функция New устанавливает соединение с базой данных
// и возвращает экземпляр БД
func New(dsn string) (*Store, error) {
	var err error

	for i := 0; i < 3; i++ {
		db, err := sqlx.ConnectContext(context.TODO(), "pgx", dsn)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		return &Store{db: db}, nil
	}

	return nil, errors.Wrap(err, "Ошибка подключения")
}

func (s Store) Posts() ([]storage.Post, error) {
	q := `
	SELECT
	p.id,
		p.title,
		p.content,
		a.id,
		a.name,
		p.created_at,
		p.updated_at
	FROM posts p
	LEFT JOIN authors a on a.id = p.author_id `

	rows, err := s.db.Query(q)
	if err != nil {
		return nil, errors.Wrap(err, "Ошибка выпролнения запроса")
	}

	var posts []storage.Post

	defer rows.Close()
	if rows.Next() {
		post := storage.Post{}
		err = rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.AuthorID,
			&post.AuthorName,
			&post.CreatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "Ошибка считывания данных")
		}
		posts = append(posts, post)
	} else {
		return nil, nil
	}

	return posts, nil
}

func (s Store) AddPost(post storage.Post) error {
	q := `
	INSERT INTO
	posts (author_id, title, content, created_at, updated_at) VALUES (:author_id, :title, :content, now(), now())
`
	rows, err := s.db.NamedQuery(q, post)
	rows.Close()
	if err != nil {
		return errors.Wrap(err, "Ошибка при добавлении поста")
	}

	return nil
}

func (s Store) UpdatePost(post storage.Post) error {
	q := `
	UPDATE posts
	SET 
	    title = $1,
	    content = $2,
	    author_id = $3,
	    updated_at = now()
	WHERE id = $4
`
	rows, err := s.db.Query(q,
		post.Title,
		post.Content,
		post.AuthorID,
		post.ID,
	)
	if err != nil {
		return errors.Wrap(err, "Ошибка при обновлении запроса")
	}
	defer rows.Close()
	return nil

}

func (s Store) DeletePost(post storage.Post) error {
	q := `DELETE FROM posts WHERE id = $1;`
	rows, err := s.db.Query(q, post.ID)
	if err != nil {
		return errors.Wrap(err, "Ошибка при удалении поста")
	}
	defer rows.Close()
	return nil
}
