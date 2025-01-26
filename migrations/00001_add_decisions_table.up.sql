CREATE TABLE user_decisions (
                                actor_user_id VARCHAR(36) NOT NULL,
                                recipient_user_id VARCHAR(36) NOT NULL,
                                liked_recipient BOOLEAN NOT NULL,
                                decision_timestamp BIGINT NOT NULL,

                                PRIMARY KEY (actor_user_id, recipient_user_id)
);

CREATE INDEX idx_liked_recipients
    ON user_decisions (recipient_user_id, decision_timestamp)
    WHERE liked_recipient = true;