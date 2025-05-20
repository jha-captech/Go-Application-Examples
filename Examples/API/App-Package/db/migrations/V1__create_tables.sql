-- Create users table
CREATE TABLE "users"
(
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT      NOT NULL,
    email      TEXT      NOT NULL UNIQUE,
    password   TEXT      NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create blogs table
CREATE TABLE "blogs"
(
    id         BIGSERIAL PRIMARY KEY,
    author_id  BIGINT    NOT NULL,
    title      TEXT      NOT NULL,
    score      REAL      NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_author
        FOREIGN KEY (author_id)
            REFERENCES "users" (id)
            ON DELETE CASCADE
);

-- Create comments table
CREATE TABLE "comments"
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT    NOT NULL,
    blog_id    BIGINT    NOT NULL,
    message    TEXT      NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_comment_user
        FOREIGN KEY (user_id)
            REFERENCES "users" (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_comment_blog
        FOREIGN KEY (blog_id)
            REFERENCES "blogs" (id)
            ON DELETE CASCADE
);

CREATE INDEX idx_comments_user_id ON "comments" (user_id);
CREATE INDEX idx_comments_blog_id ON "comments" (blog_id);
