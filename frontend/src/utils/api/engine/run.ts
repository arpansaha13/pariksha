interface EngineRunBody {
  code: string
  environment: EngineEnv
}

export async function engineRun(body: EngineRunBody) {
  const { $api } = useNuxtApp()

  return $api('/api/engine/run', {
    method: 'POST',
    body,
  })
}
