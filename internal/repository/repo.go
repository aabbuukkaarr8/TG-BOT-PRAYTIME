package repository

import (
	"gihub.com/aabbuukkaarr8/TG-BOT-PRAYTIME/internal/store"
)

type CommentRepository struct {
	store *store.Store
}

func NewCommentRepository(store *store.Store) *CommentRepository {
	return &CommentRepository{
		store: store,
	}
}
