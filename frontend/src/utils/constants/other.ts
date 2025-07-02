export enum HeaderNames {
  XCSRFToken = 'X-CSRFToken',
}

export enum CookieNames {
  CSRF_TOKEN = 'pariksha-csrftoken',
  TOKEN = 'pariksha-token',
}

export enum NuxtErrorStatusMessage {
  INCOMPLETE_EVALUATION = 'incomplete_evaluation',
}

export enum UseStateKeys {
  PreviousPath = 'previous-path',
}

export const InjectionKeys = {
  ConfirmModal: Symbol('ConfirmModal') as InjectionKey<
    ReturnType<ReturnType<typeof useOverlay>['create']>
  >,
}
