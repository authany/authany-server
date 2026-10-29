package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

// ProjectNameMaxLength is counted in characters, not bytes.
const ProjectNameMaxLength = 100

var ErrProjectNameInvalid = apierrors.Invalid.WithReason("InvalidProjectName")

// AppProfileService stores the project name: the name collaborators know a
// project by in the portal. It is unrelated to app.name, which end users see
// on the login page. (Authany)
type AppProfileService struct {
	GlobalDatabase *globaldb.Handle
	SQLBuilder     *globaldb.SQLBuilder
	SQLExecutor    *globaldb.SQLExecutor
	Clock          clock.Clock
}

// GetProjectName acquires connection. It returns nil when the project has no name.
func (s *AppProfileService) GetProjectName(ctx context.Context, appID string) (*string, error) {
	var name *string
	err := s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		q := s.SQLBuilder.Select("project_name").
			From(s.SQLBuilder.TableName("_portal_app_profile")).
			Where("app_id = ?", appID)
		row, err := s.SQLExecutor.QueryRowWith(ctx, q)
		if err != nil {
			return err
		}
		var n string
		err = row.Scan(&n)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		name = &n
		return nil
	})
	if err != nil {
		return nil, err
	}
	return name, nil
}

// GetProjectNames acquires connection. Projects without a name are absent from the map.
func (s *AppProfileService) GetProjectNames(ctx context.Context, appIDs []string) (map[string]string, error) {
	names := map[string]string{}
	if len(appIDs) == 0 {
		return names, nil
	}
	err := s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		q := s.SQLBuilder.Select("app_id", "project_name").
			From(s.SQLBuilder.TableName("_portal_app_profile")).
			Where("app_id = ANY (?)", pq.Array(appIDs))
		rows, err := s.SQLExecutor.QueryWith(ctx, q)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var appID, name string
			if err := rows.Scan(&appID, &name); err != nil {
				return err
			}
			names[appID] = name
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return names, nil
}

// SetProjectName acquires connection. An empty name removes it, so the portal
// falls back to showing the project ID.
func (s *AppProfileService) SetProjectName(ctx context.Context, appID string, name string) error {
	name, err := NormalizeProjectName(name)
	if err != nil {
		return err
	}
	return s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		if name == "" {
			q := s.SQLBuilder.Delete(s.SQLBuilder.TableName("_portal_app_profile")).
				Where("app_id = ?", appID)
			_, err := s.SQLExecutor.ExecWith(ctx, q)
			return err
		}
		q := s.SQLBuilder.Insert(s.SQLBuilder.TableName("_portal_app_profile")).
			Columns("app_id", "project_name", "updated_at").
			Values(appID, name, s.Clock.NowUTC()).
			Suffix("ON CONFLICT (app_id) DO UPDATE SET project_name = excluded.project_name, updated_at = excluded.updated_at")
		_, err := s.SQLExecutor.ExecWith(ctx, q)
		return err
	})
}

// NormalizeProjectName trims the name and rejects one that is too long or
// contains control characters.
func NormalizeProjectName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if utf8.RuneCountInString(name) > ProjectNameMaxLength {
		return "", ErrProjectNameInvalid.New("project name is too long")
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return "", ErrProjectNameInvalid.New("project name contains control characters")
		}
	}
	return name, nil
}
