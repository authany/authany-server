package graphql

import (
	"github.com/graphql-go/graphql"

	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

var updateProjectNameInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "UpdateProjectNameInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"appID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "App ID to update.",
		},
		"projectName": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The new project name. An empty string removes it.",
		},
	},
})

var updateProjectNamePayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "UpdateProjectNamePayload",
	Fields: graphql.Fields{
		"app": &graphql.Field{Type: graphql.NewNonNull(nodeApp)},
	},
})

var _ = registerMutationField(
	"updateProjectName",
	&graphql.Field{
		Description: "Update the name collaborators know the project by (Authany)",
		Type:        graphql.NewNonNull(updateProjectNamePayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(updateProjectNameInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (any, error) {
			ctx := p.Context
			input := p.Args["input"].(map[string]any)
			appNodeID := input["appID"].(string)
			projectName := input["projectName"].(string)

			resolvedNodeID := relay.FromGlobalID(appNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeApp {
				return nil, apierrors.NewInvalid("invalid app ID")
			}
			appID := resolvedNodeID.ID

			gqlCtx := GQLContext(ctx)

			// Access control: collaborator.
			_, err := gqlCtx.AuthzService.CheckAccessOfViewer(ctx, appID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.AppProfileService.SetProjectName(ctx, appID, projectName)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]any{
				"app": gqlCtx.Apps.Load(ctx, appID),
			}).Value, nil
		},
	},
)
