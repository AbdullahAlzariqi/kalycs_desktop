import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'

// Base API slice. Use api.injectEndpoints in your feature files to extend this.
export const api = createApi({
    reducerPath: 'api',
    baseQuery: fetchBaseQuery({
        baseUrl: '/api',
        // prepareHeaders: (headers, { getState }) => {
        //     // Example: add auth token
        //     // const token = selectAuthToken(getState())
        //     // if (token) headers.set('authorization', `Bearer ${token}`)
        //     return headers
        // },
    }),
    tagTypes: [], // Add tag types as you create endpoints
    endpoints: () => ({}),
})

// You can export hooks from injected endpoints like so:
// export const { useGetSomethingQuery } = api 