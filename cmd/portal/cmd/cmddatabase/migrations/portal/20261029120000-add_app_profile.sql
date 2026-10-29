-- +migrate Up
CREATE TABLE _portal_app_profile
(
    app_id       text PRIMARY KEY,
    project_name text NOT NULL,
    updated_at   timestamp without time zone NOT NULL
);

-- +migrate Down
DROP TABLE _portal_app_profile;
