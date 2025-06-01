import type { Brand } from 'ts-brand'

type StoreId = Brand<string, 'store'>

export const authStoreId = 'auth-store-id' as StoreId
export const newExamStoreId = 'new-exam-store-id' as StoreId
export const editorStoreId = 'editor-store-id' as StoreId
export const paperStoreId = 'paper-store-id' as StoreId
