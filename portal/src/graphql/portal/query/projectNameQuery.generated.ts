import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type ProjectNameQueryQueryVariables = Types.Exact<{
  id: Types.Scalars['ID']['input'];
}>;


export type ProjectNameQueryQuery = { __typename?: 'Query', node?: { __typename: 'App', id: string, projectName?: string | null } | { __typename: 'User' } | { __typename: 'Viewer' } | null };


export const ProjectNameQueryDocument = gql`
    query projectNameQuery($id: ID!) {
  node(id: $id) {
    __typename
    ... on App {
      id
      projectName
    }
  }
}
    `;

/**
 * __useProjectNameQueryQuery__
 *
 * To run a query within a React component, call `useProjectNameQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useProjectNameQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useProjectNameQueryQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useProjectNameQueryQuery(baseOptions: Apollo.QueryHookOptions<ProjectNameQueryQuery, ProjectNameQueryQueryVariables> & ({ variables: ProjectNameQueryQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<ProjectNameQueryQuery, ProjectNameQueryQueryVariables>(ProjectNameQueryDocument, options);
      }
export function useProjectNameQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<ProjectNameQueryQuery, ProjectNameQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<ProjectNameQueryQuery, ProjectNameQueryQueryVariables>(ProjectNameQueryDocument, options);
        }
// @ts-ignore
export function useProjectNameQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<ProjectNameQueryQuery, ProjectNameQueryQueryVariables>): Apollo.UseSuspenseQueryResult<ProjectNameQueryQuery, ProjectNameQueryQueryVariables>;
export function useProjectNameQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<ProjectNameQueryQuery, ProjectNameQueryQueryVariables>): Apollo.UseSuspenseQueryResult<ProjectNameQueryQuery | undefined, ProjectNameQueryQueryVariables>;
export function useProjectNameQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<ProjectNameQueryQuery, ProjectNameQueryQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<ProjectNameQueryQuery, ProjectNameQueryQueryVariables>(ProjectNameQueryDocument, options);
        }
export type ProjectNameQueryQueryHookResult = ReturnType<typeof useProjectNameQueryQuery>;
export type ProjectNameQueryLazyQueryHookResult = ReturnType<typeof useProjectNameQueryLazyQuery>;
export type ProjectNameQuerySuspenseQueryHookResult = ReturnType<typeof useProjectNameQuerySuspenseQuery>;
export type ProjectNameQueryQueryResult = Apollo.QueryResult<ProjectNameQueryQuery, ProjectNameQueryQueryVariables>;