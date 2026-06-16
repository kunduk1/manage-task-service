CREATE INDEX idx_task_comments_task ON task_comments (task_id);
DROP INDEX idx_task_comments_task_created ON task_comments;

CREATE INDEX idx_task_history_task ON task_history (task_id);
DROP INDEX idx_task_history_task_changed ON task_history;

DROP INDEX idx_tasks_created_at_team_creator ON tasks;
