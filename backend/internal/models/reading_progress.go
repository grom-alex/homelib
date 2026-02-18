package models

import "time"

type ReadingProgress struct {
	ID              int64     `json:"-"`
	UserID          string    `json:"-"`
	BookID          int64     `json:"-"`
	ChapterID       string    `json:"chapterId"`
	ChapterProgress int       `json:"chapterProgress"`
	TotalProgress   int       `json:"totalProgress"`
	Device          string    `json:"device"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type SaveProgressInput struct {
	ChapterID       string `json:"chapterId" binding:"required"`
	ChapterProgress int    `json:"chapterProgress" binding:"min=0,max=100"`
	TotalProgress   int    `json:"totalProgress" binding:"min=0,max=100"`
	Device          string `json:"device"`
}
