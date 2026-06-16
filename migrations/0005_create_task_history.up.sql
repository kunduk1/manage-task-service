CREATE TABLE IF NOT EXISTS task_history (
    id         BIGINT      NOT NULL AUTO_INCREMENT,
    task_id    BIGINT      NOT NULL,
    changed_by BIGINT      NOT NULL,
    field      VARCHAR(64) NOT NULL,
    old_value  TEXT        NULL,
    new_value  TEXT        NULL,
    changed_at TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_task_history_task (task_id),
    KEY idx_task_history_changed_by (changed_by),
    CONSTRAINT fk_task_history_task       FOREIGN KEY (task_id)    REFERENCES tasks (id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    CONSTRAINT fk_task_history_changed_by FOREIGN KEY (changed_by) REFERENCES users (id) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
