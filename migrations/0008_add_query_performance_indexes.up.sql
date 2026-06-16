-- tasks: покрывающий индекс для лидерборда «топ создателей» (TopCreators):
-- диапазон по created_at + группировка по (team_id, created_by) без обращения к таблице.
CREATE INDEX idx_tasks_created_at_team_creator ON tasks (created_at, team_id, created_by);

-- task_history: составной индекс под ListByTask (фильтр task_id + сортировка changed_at, id);
-- заменяет одиночный idx_task_history_task — FK fk_task_history_task остаётся покрытым
-- ведущим столбцом task_id составного индекса.
CREATE INDEX idx_task_history_task_changed ON task_history (task_id, changed_at, id);
DROP INDEX idx_task_history_task ON task_history;

-- task_comments: составной индекс под ListByTask (фильтр task_id + сортировка created_at, id);
-- заменяет одиночный idx_task_comments_task — FK fk_task_comments_task остаётся покрытым.
CREATE INDEX idx_task_comments_task_created ON task_comments (task_id, created_at, id);
DROP INDEX idx_task_comments_task ON task_comments;
