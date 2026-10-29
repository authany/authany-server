import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type UpdateProjectNameMutationMutationVariables = Types.Exact<{
  appID: Types.Scalars['ID']['input'];
  projectName: Types.Scalars['String']['input'];
}>;


export type UpdateProjectNameMutationMutation = { __typename?: 'Mutation', updateProjectName: { __typename?: 'UpdateProjectNamePayload', app: { __typename?: 'App', id: string, projectName?: string | null } } };


export const UpdateProjectNameMutationDocument = gql`
    mutation updateProjectNameMutation($appID: ID!, $projectName: String!) {
  updateProjectName(input: {appID: $appID, projectName: $projectName}) {
    app {
      id
      projectName
    }
  }
}
    `;
export type UpdateProjectNameMutationMutationFn = Apollo.MutationFunction<UpdateProjectNameMutationMutation, UpdateProjectNameMutationMutationVariables>;

/**
 * __useUpdateProjectNameMutationMutation__
 *
 * To run a mutation, you first call `useUpdateProjectNameMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateProjectNameMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateProjectNameMutationMutation, { data, loading, error }] = useUpdateProjectNameMutationMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *      projectName: // value for 'projectName'
 *   },
 * });
 */
export function useUpdateProjectNameMutationMutation(baseOptions?: Apollo.MutationHookOptions<UpdateProjectNameMutationMutation, UpdateProjectNameMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateProjectNameMutationMutation, UpdateProjectNameMutationMutationVariables>(UpdateProjectNameMutationDocument, options);
      }
export type UpdateProjectNameMutationMutationHookResult = ReturnType<typeof useUpdateProjectNameMutationMutation>;
export type UpdateProjectNameMutationMutationResult = Apollo.MutationResult<UpdateProjectNameMutationMutation>;
export type UpdateProjectNameMutationMutationOptions = Apollo.BaseMutationOptions<UpdateProjectNameMutationMutation, UpdateProjectNameMutationMutationVariables>;