package models

type Task struct {
	Level       int8
	Title       string
	Description string
}

type TemplateRepo struct {
	Name         string
	Owner        string
	Organization string
}

type Challenge struct {
	Name   string
	TeamID string
	Repo   TemplateRepo
	Tasks  []Task
}
