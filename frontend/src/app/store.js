import { configureStore } from '@reduxjs/toolkit'
import { api } from './api'

// Add additional slice reducers here as you create them
export const store = configureStore({
    reducer: {
        [api.reducerPath]: api.reducer,
    },
    middleware: (getDefaultMiddleware) =>
        getDefaultMiddleware().concat(api.middleware),
})

export default store 